import os
import subprocess
import yaml
from datetime import datetime
import pytz
import argparse

MNT_COLLECTOR_PATH = "/home/exu03/workspace/mnt-collector/mnt-collector"

def log_event(log_file, event_time, event_type, status, source_yaml, command):
    """Log an event to the log file."""
    with open(log_file, 'a') as f:
        f.write(f"{event_time},{event_type},{status},{source_yaml},\"{command}\"\n")

def read_yaml_cases(yaml_path):
    """Read the cases from the YAML file."""
    with open(yaml_path, 'r') as f:
        data = yaml.safe_load(f)
    return data.get('cases', [])

def generate_simulations_yaml(simulations_path, title_path, traceid_files):
    """Generate the simulations YAML file."""
    with open(title_path, 'r') as f:
        title_content = f.read()
    
    trace_ids = []
    for traceid_file in traceid_files:
        with open(traceid_file, 'r') as f:
            trace_ids.extend(f.readlines())
    
    with open(simulations_path, 'w') as f:
        f.write(title_content)
        f.write("\ntrace-id:\n")
        f.writelines(trace_ids)

def run_command(command, log_file, event_type, source_yaml):
    """Run a command and log its start and end."""
    est = pytz.timezone('US/Eastern')
    start_time = datetime.now(est).strftime('%Y-%m-%d_%H:%M:%S.%f')
    log_event(log_file, start_time, event_type, "start", source_yaml, command)
    
    result = subprocess.run(command, shell=True)
    
    end_time = datetime.now(est).strftime('%Y-%m-%d_%H:%M:%S.%f')
    status = "end" if result.returncode == 0 else "error"
    log_event(log_file, end_time, event_type, status, source_yaml, command)
    
    if result.returncode != 0:
        raise RuntimeError(f"Command failed: {command}")

def main():
    parser = argparse.ArgumentParser(description="Run simulation tasks.")
    parser.add_argument("--collect", required=True, help="Path to the YAML file.")
    args = parser.parse_args()

    yaml_path = args.collect

    base_dir = os.path.dirname(yaml_path)
    yaml_file_name = os.path.basename(yaml_path)
    log_file = os.path.join(base_dir, yaml_file_name.replace(".yaml", "-log.csv"))
    simulations_path = os.path.join(base_dir, yaml_file_name.replace(".yaml", "-simulations.yaml"))
    title_path = os.path.join(base_dir, "../simulations-title.yaml")
    
    # Step 1: Run profiles command
    profiles_command = f"{MNT_COLLECTOR_PATH} profiles --collect {yaml_path}"
    run_command(profiles_command, log_file, "profiles", yaml_path)
    
    # Step 2: Run traces command
    traces_command = f"{MNT_COLLECTOR_PATH} traces --collect {yaml_path}"
    run_command(traces_command, log_file, "traces", yaml_path)
    
    # Step 3: Read traceid files
    cases = read_yaml_cases(yaml_path)
    traceid_files = []
    for case in cases:
        suite = case['suite']
        title = case['title']
        traceid_file = os.path.join(base_dir, f"../../traceid/{suite}-{title}.txt")
        traceid_files.append(traceid_file)
    
    # Step 4: Generate simulations YAML
    generate_simulations_yaml(simulations_path, title_path, traceid_files)
    
    # Step 5: Run simulations command
    simulations_command = f"{MNT_COLLECTOR_PATH} simulations --collect {simulations_path}"
    run_command(simulations_command, log_file, "simulations", simulations_path)

if __name__ == "__main__":
    main()