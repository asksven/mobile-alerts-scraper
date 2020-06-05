#!/bin/bash

for FILE in `ls -tU $LOGDIR/*.log`
do
   echo "Processing file $FILE"
   python logparser.py $FILE
   mv $FILE $FILE.processed
done
