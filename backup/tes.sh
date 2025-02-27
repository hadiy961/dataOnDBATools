#!/bin/bash

###########################################
# Configuration Section
###########################################
readonly CONFIG=(
    MYSQL_USER="sst_user"
    MYSQL_PASS="demo"
    MYSQL_REPL_USER="maxscale"
    MYSQL_REPL_PASS="F33D9A6E1376BD25313EB4EF0733ED43"
    COMPRESSION_TOOL="pigz"
    COMPRESSION_THREADS="4"
)

# Load configuration
for config in "${CONFIG[@]}"; do
    declare "$config"
done

###########################################
# Constants
###########################################
readonly TIMESTAMP_FORMAT='+%Y-%m-%d %H:%M:%S'
readonly SYSTEM_DBS=('information_schema' 'mysql' 'performance_schema' 'sys')
readonly SYSTEM_DBS_EXCLUSION="WHERE \`Database\` NOT IN ('${SYSTEM_DBS[*]}')"

###########################################
# Directory Setup
###########################################
setup_directories() {
    local date_dir=$(date +%Y%m%d)
    readonly BASE_DIR="./${date_dir}"
    readonly LOGS_DIR="${BASE_DIR}/logs"
    readonly DATA_DIR="${BASE_DIR}/data"
    readonly LOG_FILE="${LOGS_DIR}/backup_$(date +%Y%m%d_%H%M%S).log"

    mkdir -p "${LOGS_DIR}" "${DATA_DIR}"
    
    # Configure logging
    exec 1> >(tee -a "${LOG_FILE}")
    exec 2>&1
}

###########################################
# Logging Functions
###########################################
log() {
    local level=$1
    local message=$2
    local timestamp=$(date "${TIMESTAMP_FORMAT}")
    echo "[${timestamp}] ${level}: ${message}"
}

log_error() { log "❌ ERROR" "$1"; }
log_success() { log "✅ SUCCESS" "$1"; }
log_info() { log "ℹ️ INFO" "$1"; }
log_warning() { log "⚠️ WARNING" "$1"; }

###########################################
# MySQL Helper Functions
###########################################
get_mysql_connection() {
    local host=$1
    local port=$2
    echo "-u${MYSQL_USER} -p${MYSQL_PASS} -h${host} -P${port}"
}

execute_mysql_query() {
    local connection=$1
    local query=$2
    local error_msg=${3:-"MySQL query failed"}
    
    mysql ${connection} -N -e "${query}" 2>/dev/null || handle_error "${error_msg}"
}

get_databases() {
    local connection=$1
    execute_mysql_query "${connection}" "SELECT SCHEMA_NAME FROM information_schema.schemata WHERE SCHEMA_NAME NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')"
}

###########################################
# Error Handling
###########################################
handle_error() {
    local exit_code=$?
    local error_msg="$1"
    log_error "${error_msg} (Exit code: ${exit_code})"
    exit ${exit_code}
}

trap 'handle_error "An unexpected error occurred"' ERR

###########################################
# Validation Functions
###########################################
validate_inputs() {
    if [ $# -ne 2 ]; then
        handle_error "Please provide HOST and PORT"
    fi

    local host=$1
    local port=$2

    # Validate port
    if ! [[ $port =~ ^[0-9]+$ ]] || [ $port -lt 1 ] || [ $port -gt 65535 ]; then
        handle_error "Invalid port number: $port"
    fi
    
    # Validate host format
    if ! [[ $host =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]] && 
       ! [[ $host =~ ^[a-zA-Z0-9][-a-zA-Z0-9.]+$ ]]; then
        handle_error "Invalid host format: $host"
    fi
}

###########################################
# Core Functions
###########################################
check_databases() {
   log_info "Checking Local Databases"
   local databases
   local output_file="${BASE_DIR}/db_list_found.txt"
   databases=$(get_databases "")
   
   if [ -z "$databases" ]; then
       log_info "No user databases found"
       echo "No user databases found" > "$output_file"
   else
       echo "$databases" > "$output_file"
       log_info "Found non-system databases: $output_file"
   fi

   log_success "Database check completed"
}

drop_all_databases() {
    log_info "Checking for non-system databases"
    local databases
    # The issue: empty connection parameter means local connection
    databases=$(get_databases "") 
    
    if [ -z "$databases" ]; then
        log_info "No user databases found to drop"
        return 0
    fi
    
    log_info "Starting to drop all non-system databases"
    local count=0
    for db in $databases; do
        log_info "Dropping database: $db"
        execute_mysql_query "" "DROP DATABASE \`$db\`" "Failed to drop database: $db"
        ((count++))
    done
    log_success "Successfully dropped $count databases"
}

backup_user_grants() {
    local connection=$1
    local output_file="${BASE_DIR}/user_grant_$(date +%Y%m%d_%H%M%S).sql"
    
    log_info "Backing up remote user grants"
    
    {
        execute_mysql_query "${connection}" "
            SELECT CONCAT('SHOW GRANTS FOR ''',user,'''@''',host,''';')
            FROM mysql.user WHERE user<>''" | \
        execute_mysql_query "${connection}" "source" | \
        sed 's/$/;/g' > "$output_file"
        
        echo "FLUSH PRIVILEGES;" >> "$output_file"
    } || handle_error "Failed to backup user grants"
    
    local grant_count=$(grep -c "GRANT" "$output_file")
    log_success "Backed up $grant_count user grants to $output_file"
}

format_database_list() {
    local db_list=("$@")
    local max_dbs_per_line=4
    local total_dbs=${#db_list[@]}
    local current_count=0
    local output=""
    
    for db in "${db_list[@]}"; do
        if [ $current_count -eq 0 ]; then
            output+="${db}"
        else
            output+=",$db"
        fi
        
        ((current_count++))
        
        if [ $current_count -eq $max_dbs_per_line ] && [ $current_count -lt $total_dbs ]; then
            output+=$'\n'
            current_count=0
        fi
    done
    
    echo "$output"
}

# Add this function after MySQL Helper Functions section
set_statement_timeout() {
   local timeout=$1
   local message="Setting max_statement_time to $timeout"
   log_info "$message"
   mysql -u$MYSQL_USER -p$MYSQL_PASS -e "set global max_statement_time = $timeout;"
}

backup_databases() {
    local host=$1
    local port=$2
    local connection=$(get_mysql_connection "$host" "$port")
    
    # Debug connection before backup
    log_info "Testing connection to remote server..."
    if ! mysql ${connection} -e "SELECT VERSION()" > /dev/null 2>&1; then
        log_error "Cannot connect to remote MySQL server"
        return 1
    fi
    
    local dblist
    readarray -t db_array < <(get_databases "${connection}")
    if [ ${#db_array[@]} -eq 0 ]; then
        log_warning "No non-system databases found to backup"
        return 0
    fi
    # Total database count
    local total_databases=${#db_array[@]}
    log_info "Total databases to backup: ${total_databases}"

    # Format database list for logging
    local formatted_dblist=$(format_database_list "${db_array[@]}")
    log_info "Preparing to backup databases: $formatted_dblist"
    
    # Setup compression
    local compress_cmd
    if [ "$COMPRESSION_TOOL" = "pigz" ] && command -v pigz >/dev/null 2>&1; then
        compress_cmd="pigz -p $COMPRESSION_THREADS"
        log_info "Using pigz compression with $COMPRESSION_THREADS threads"
    else
        if [ "$COMPRESSION_TOOL" = "pigz" ]; then
            log_warning "pigz not installed. Falling back to gzip"
        fi
        compress_cmd="gzip"
        log_info "Using gzip compression"
    fi
    
    # Set statement timeout to 0 to prevent interruptions during large database backups
    set_statement_timeout 0
    
    # Create a single backup for all databases
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local output_file="${DATA_DIR}/backup_all_databases_${timestamp}.sql.gz"
    local error_file="${LOGS_DIR}/mysqldump_all_databases_error.log"
    local progress_file="${LOGS_DIR}/backup_progress_${timestamp}.tmp"
    
    # Debug: Show exact mysqldump command for all databases
    # Create mysqldump command that includes all databases
    local mysqldump_cmd="mysqldump ${connection} --databases ${db_array[@]} -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=true --triggers=true"
    
    log_info "[Backup][Starting] All databases (${total_databases} total)"
    
    # Create an empty output file so we can monitor its size
    touch "${output_file}"
    
    # Start the progress monitoring in background
    (
        local last_size=0
        local start_time=$(date +%s)
        
        while true; do
            sleep 1  # Check every 10 seconds
            
            # Exit if progress file is removed
            if [ ! -f "${progress_file}" ]; then
                break
            fi
            
            if [ -f "${output_file}" ]; then
                local current_size=$(stat -c %s "${output_file}" 2>/dev/null || echo "0")
                local elapsed=$(($(date +%s) - start_time))
                # Calculate MB without using bc (divide by 1048576)
                local size_mb=$(awk "BEGIN {printf \"%.2f\", ${current_size}/1048576}")
                
                # If size changed, update progress on single line
                if [ "${current_size}" != "${last_size}" ]; then
                    # Calculate speed in MB/s
                    local size_diff=$((current_size - last_size))
                    local speed=0
                    if [ $elapsed -gt 0 ]; then
                        speed=$(awk "BEGIN {printf \"%.2f\", ${size_diff}/1048576/10}")
                    fi
                    
                    # Use carriage return to update the same line
                    printf "\r[Backup][Progress] Size: %s MB, Speed: %s MB/s, Time: %ss" "${size_mb}" "${speed}" "${elapsed}"
                    last_size="${current_size}"
                fi
            fi
        done
    ) &
    
    # Create progress file to signal monitoring process
    touch "${progress_file}"
    
    # Run the backup command
    set -o pipefail
    if ! eval "$mysqldump_cmd" 2> >(tee "$error_file" >&2) | ${compress_cmd} > "${output_file}"; then
        # Stop progress monitoring
        rm -f "${progress_file}"
        
        # Check if error file has content
        if [ -s "$error_file" ]; then
            log_error "MySQL Error Details: $(cat "$error_file")"
        else
            # Remove empty error file
            rm -f "$error_file"
        fi
        handle_error "[Backup][Failed] All databases backup"
        # Reset statement timeout
        set_statement_timeout 60
        return 1
    else
        # Remove error log file if backup was successful
        rm -f "$error_file"
    fi
    
    # Stop progress monitoring
    rm -f "${progress_file}"
    
    # Check backup status
    if [ ! -f "${output_file}" ] || [ ! -s "${output_file}" ]; then
        handle_error "[Backup][Empty] All databases backup"
        # Reset statement timeout
        set_statement_timeout 60
        return 1
    fi

    local backup_size=$(du -h "${output_file}" | cut -f1)
    log_info ""
    log_info "[Backup][Complete] All databases (${backup_size})"
    
    # Reset statement timeout
    set_statement_timeout 60
    
    # Summary log
    log_success "Backup Process Complete: All ${total_databases} databases backed up in a single file"
    
    return 0
}

get_binlog_info() {
    local connection=$1
    local output_file="${BASE_DIR}/gtid_binlog_$(date +%Y%m%d_%H%M%S).txt"
    
    log_info "Retrieving binlog information"
    
    local start_binlog start_pos gtid_binlog
    start_binlog=$(execute_mysql_query "${connection}" "SHOW MASTER STATUS" | awk '{print $1}')
    start_pos=$(execute_mysql_query "${connection}" "SHOW MASTER STATUS" | awk '{print $2}')
    gtid_binlog=$(execute_mysql_query "${connection}" "SELECT BINLOG_GTID_POS('${start_binlog}', ${start_pos})")
    
    # Changed format to remove spaces around = sign
    cat > "$output_file" << EOF || handle_error "Failed to save binlog information"
MASTER_LOG_FILE=$start_binlog
MASTER_LOG_POS=$start_pos
gtid_binlog=$gtid_binlog
EOF

    log_success "Binlog information saved to $output_file"
}

restore_user_grants() {
    local latest_grant=$(ls -t ${BASE_DIR}/user_grant_*.sql 2>/dev/null | head -1)
    
    if [ -f "$latest_grant" ]; then
        log_info "Restoring user grants from $latest_grant"
        execute_mysql_query "" "source ${latest_grant}" "Failed to restore user grants"
        log_success "User grants restored successfully"
    else
        log_warning "No grant file found to restore"
        return 0
    fi
}

configure_replication() {
    local host=$1
    local port=$2
    local latest_binlog=$(ls -t ${BASE_DIR}/gtid_binlog_*.txt 2>/dev/null | head -1)
    
    if [ ! -f "$latest_binlog" ]; then
        handle_error "No binlog info file found"
    fi

    log_info "Configuring replication using $latest_binlog"
    
    # Read variables from file without executing them
    MASTER_LOG_FILE=$(grep "^MASTER_LOG_FILE=" "$latest_binlog" | cut -d'=' -f2)
    MASTER_LOG_POS=$(grep "^MASTER_LOG_POS=" "$latest_binlog" | cut -d'=' -f2)
    gtid_binlog=$(grep "^gtid_binlog=" "$latest_binlog" | cut -d'=' -f2)
    
    log_info "Executing replication commands..."
    mysql -u${MYSQL_USER} -p${MYSQL_PASS} << EOF
        STOP SLAVE;
        RESET SLAVE;
        SET GLOBAL gtid_slave_pos='$gtid_binlog';
        CHANGE MASTER TO
            MASTER_HOST='$host',
            MASTER_PORT=$port,
            MASTER_USER='$MYSQL_REPL_USER',
            MASTER_PASSWORD='$MYSQL_REPL_PASS',
            MASTER_USE_GTID=slave_pos;
        START SLAVE;
EOF
    
    if [ $? -eq 0 ]; then
        log_success "Replication configured successfully"
    else
        handle_error "Failed to configure replication"
    fi
}

check_replication() {
   local max_attempts=60  # Maximum number of attempts (1 minute with 1-second intervals)
   local attempt=1
   local status_ok=false
   
   log_info "Checking replication status"
   
   while [ $attempt -le $max_attempts ]; do
       # Get replication status
       local slave_status
       slave_status=$(mysql -u${MYSQL_USER} -p${MYSQL_PASS} -e "SHOW SLAVE STATUS\G")
       
       # Extract required values
       local io_running=$(echo "$slave_status" | grep 'Slave_IO_Running:' | awk '{print $2}')
       local sql_running=$(echo "$slave_status" | grep 'Slave_SQL_Running:' | awk '{print $2}')
       local seconds_behind=$(echo "$slave_status" | grep 'Seconds_Behind_Master:' | awk '{print $2}')
       
       # Log current status
       log_info "Attempt $attempt: IO_Running=$io_running, SQL_Running=$sql_running, Seconds_Behind=$seconds_behind"
       
       # Check if all conditions are met
       if [ "$io_running" = "Yes" ] && [ "$sql_running" = "Yes" ] && [ "$seconds_behind" = "0" ]; then
           status_ok=true
           break
       fi
       
       # If not met, check for specific issues
       if [ "$io_running" != "Yes" ]; then
           local io_error=$(echo "$slave_status" | grep 'Last_IO_Error:' | cut -d: -f2-)
           log_warning "IO Thread not running. Error: $io_error"
       fi
       
       if [ "$sql_running" != "Yes" ]; then
           local sql_error=$(echo "$slave_status" | grep 'Last_SQL_Error:' | cut -d: -f2-)
           log_warning "SQL Thread not running. Error: $sql_error"
       fi
       
       if [ "$seconds_behind" != "0" ]; then
           log_info "Slave is $seconds_behind seconds behind master. Waiting for catch up..."
       fi
       
       # Wait before next attempt
       sleep 1
       ((attempt++))
   done
   
   # Final check
   if [ "$status_ok" = true ]; then
       log_success "Replication is running correctly and synchronized"
       return 0
   else
       handle_error "Replication check failed after $max_attempts attempts"
       return 1
   fi
}

restore_database() {
    local host=$1
    local port=$2
    local connection=$(get_mysql_connection "$host" "$port")
    
    # Debug connection before restore
    log_info "Testing connection to remote server..."
    if ! mysql -u$MYSQL_USER -p$MYSQL_PASS  -e "SELECT VERSION()" > /dev/null 2>&1; then
        log_error "Cannot connect to remote MySQL server"
        return 1
    fi
    
    # Find latest backup file
    local latest_backup=$(ls -t ${DATA_DIR}/backup_all_databases_*.sql.gz 2>/dev/null | head -1)
    
    if [ -z "$latest_backup" ]; then
        handle_error "No backup files found to restore"
        return 1
    fi
    
    # Extract timestamp from filename
    local timestamp=$(echo "$latest_backup" | sed -n 's/.*backup_all_databases_\([0-9]\{8\}_[0-9]\{6\}\)\.sql\.gz/\1/p')
    local backup_size=$(du -h "$latest_backup" | cut -f1)
    
    log_info "Using latest backup from $timestamp ($backup_size)"
    
    # Set statement timeout to 0 to prevent interruptions during restore
    set_statement_timeout 0
    
    # Check if backup file exists and is not empty
    if [ ! -s "$latest_backup" ]; then
        handle_error "[Restore][Invalid] Backup file is empty or invalid"
        set_statement_timeout 60
        return 1
    fi
    
    # Check if mysql data directory exists
    local mysql_data_dir="/var/lib/mysql"
    if [ ! -d "$mysql_data_dir" ]; then
        log_warning "MySQL data directory not found at $mysql_data_dir, size monitoring disabled"
        mysql_data_dir=""
    else
        # Get initial size of MySQL data directory
        local initial_size=$(du -sh $mysql_data_dir 2>/dev/null | awk '{print $1}')
        log_info "Initial MySQL data directory size: $initial_size"
    fi
    
    # Prepare progress monitoring
    local progress_file="${LOGS_DIR}/restore_progress_${timestamp}.tmp"
    local error_file="${LOGS_DIR}/mysql_restore_error.log"
    
    # Start the restore process
    log_info "[Restore][Starting] All databases from backup $timestamp ($backup_size)"
    
    # Create progress file to signal monitoring process
    touch "$progress_file"
    
    # Start monitoring in background with improved formatting similar to backup_database
    (
        local start_time=$(date +%s)
        local last_db=""
        local current_db=""
        local last_size=0
        
        # Display progress during restore
        while [ -f "$progress_file" ]; do
            sleep 1
            local elapsed=$(($(date +%s) - start_time))
            
            # Get current MySQL data directory size
            if [ -n "$mysql_data_dir" ] && [ -d "$mysql_data_dir" ]; then
                local current_size_bytes=$(du -sb $mysql_data_dir 2>/dev/null | awk '{print $1}')
                local current_size_mb=$(awk "BEGIN {printf \"%.2f\", ${current_size_bytes}/1048576}")
                
                # Calculate speed in MB/s if we have previous measurements
                local speed="0.00"
                if [ $last_size -gt 0 ] && [ $elapsed -gt 0 ]; then
                    local size_diff=$((current_size_bytes - last_size))
                    speed=$(awk "BEGIN {printf \"%.2f\", ${size_diff}/1048576}")
                fi
                last_size=$current_size_bytes
                
                # Try to get the most recently created database
                if mysql -e "SHOW DATABASES" &>/dev/null; then
                    # Get list of databases sorted by modification time of their directory
                    current_db=$(find $mysql_data_dir -maxdepth 1 -type d -not -name "mysql" -not -name "performance_schema" \
                        -not -name "information_schema" -not -name "sys" -not -name "." -not -name ".." \
                        -printf "%T@ %f\n" 2>/dev/null | sort -nr | head -1 | awk '{print $2}')
                    
                    # If we found a new database, update last_db
                    if [ -n "$current_db" ] && [ "$current_db" != "$last_db" ]; then
                        last_db=$current_db
                    fi
                fi
                
                # Use carriage return to update the same line
                printf "\r[Restore][Progress] Size: %s MB, Speed: %s MB/s, Time: %ss, Latest DB: %s" "${current_size_mb}" "${speed}" "${elapsed}" "${last_db}"
            else
                # If we can't monitor directory size, just show elapsed time
                printf "\r[Restore][Progress] Time: %ss, Latest DB: %s" "${elapsed}" "${last_db}"
            fi
        done
        echo "" # Add newline at end
    ) &
    local monitor_pid=$!
    
    # Start the restore process
    set -o pipefail
    if ! gunzip -c "$latest_backup" | mysql -u$MYSQL_USER -p$MYSQL_PASS -f 2> >(tee "$error_file" >&2); then
        # Stop progress monitoring
        rm -f "$progress_file"
        wait $monitor_pid 2>/dev/null || true
        
        if [ -s "$error_file" ]; then
            log_error "MySQL Restore Error Details: $(cat "$error_file")"
        else
            # Remove empty error file
            rm -f "$error_file"
        fi
        handle_error "[Restore][Failed] All databases restore"
        set_statement_timeout 60
        return 1
    else
        # Remove error log file if restore was successful
        rm -f "$error_file"
    fi
    
    # Stop progress monitoring
    rm -f "$progress_file"
    wait $monitor_pid 2>/dev/null || true
    
    # Get final MySQL data directory size
    if [ -n "$mysql_data_dir" ] && [ -d "$mysql_data_dir" ]; then
        local final_size=$(du -sh $mysql_data_dir 2>/dev/null | awk '{print $1}')
        log_info "Final MySQL data directory size: $final_size"
    fi
    
    log_info "[Restore][Complete] All databases restored ($backup_size)"
    
    # Reset statement timeout
    set_statement_timeout 60
    
    # Summary log
    log_success "Restore Process Complete: All databases restored from backup"
    
    return 0
}

###########################################
# Main Function
###########################################
main() {
    clear
    validate_inputs "$@"
    setup_directories
    
    local host=$1
    local port=$2
    local connection=$(get_mysql_connection "$host" "$port")
    
    log_info "Starting MySQL operations for host: $host:$port"

    # Execute operations with progress tracking
    local total_steps=7
    local current_step=0

    ((current_step++))
    log_info "[$current_step/$total_steps] Checking databases"
    check_databases

    ((current_step++))
    log_info "[$current_step/$total_steps] Dropping existing databases"
    drop_all_databases

    ((current_step++))
    log_info "[$current_step/$total_steps] Backing up user grants"
    backup_user_grants "$connection"

    ((current_step++))
    log_info "[$current_step/$total_steps] Restoring user grants"
    restore_user_grants

    ((current_step++))
    log_info "[$current_step/$total_steps] Getting binlog information"
    get_binlog_info "$connection"

    ((current_step++))
    log_info "[$current_step/$total_steps] Backing up databases"
    backup_databases "$host" "$port"

    ((current_step++))
    log_info "[$current_step/$total_steps] Restoring database"
    restore_database

    ((current_step++))
    log_info "[$current_step/$total_steps] Configuring replication"
    configure_replication "$host" "$port"

    ((current_step++))
    log_info "[$current_step/$total_steps] Verifying replication"
    check_replication

    echo "================================"
    log_success "All operations completed successfully!"
}

# Execute main function
main "$@"