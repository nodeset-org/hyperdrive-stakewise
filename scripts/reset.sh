#!/usr/bin/env bash 

echo "Cleaning up previous configuration..."
    
remove_containers

# clear old data (if any)
rm -rd $data_dir/nimbus-data
rm -rd $data_dir/tmp
rm -rd $data_dir/stakewise-data
rm -rd $data_dir/geth-data