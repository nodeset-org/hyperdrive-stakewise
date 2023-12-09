#!/usr/bin/env bash 

echo "Cleaning up previous configuration..."
    
"$SCRIPT_DIR/nodeset.sh" shutdown

# clear old data (if any)
sudo rm -rd $data_dir/nimbus-data
sudo rm -rd $data_dir/tmp
sudo rm -rd $data_dir/stakewise-data
sudo rm -rd $data_dir/geth-data