# Query Analysis (Beta)

## Objective and scope
This tool examines the structure of workload queries. This analysis serves both analysts and engineers, providing valuable insights. Furthermore, it can facilitate the simulation of user query workloads.

## How to run the tool
```
python3 query_analysis.py --test_database <querylog_database>
```

## Prerequisites 
- python3 should be installed on the system
- Need input database in sqlite format that has a table called slowlog. slowlog table
  can be constructed from slowlog using O11 tool. 
- Need to install sqlite3 since the tool uses it to access the log data. 

## Sample test
We have a sample test of 10,000 queries in sample_data/slowlog.sql in sqlite format.
You can create a sqlite test database from the sample data and then run the tool on it.
 
```
sqlite3 testDB < sample_data/slowlog.sql 
python3 query_analysis.py --test_database sample_data/testDB
```

## Sample result
The result from the sample test above is listed below. UNKNOWN represents the cases
we could not parse the SQL. Further work is needed to fix the logging code.
```
Query Type                          Frequency 
SCAN                                3280      
INSERT_VALUES                       3209      
AGGREGATE_SCAN_NO_GROUPBY           1477      
AGGREGATE_JOIN_GROUPBY              732       
UNKNOWN                             407       
DELETE                              355       
AGGREGATE_SCAN_GROUPBY              254       
SYSTEM                              220       
AGGREGATE_JOIN_NO_GROUPBY           54        
UPDATE                              10        
JOIN_NO_AGGREGATE                   2         


Query Type                          Total Query Time in Seconds
AGGREGATE_SCAN_NO_GROUPBY           5940.73             
UNKNOWN                             1216.44             
INSERT_VALUES                       1174.3              
AGGREGATE_JOIN_GROUPBY              930.78              
SCAN                                380.37              
DELETE                              345.67              
AGGREGATE_SCAN_GROUPBY              336.01              
AGGREGATE_JOIN_NO_GROUPBY           2.72                
JOIN_NO_AGGREGATE                   0.78                
SYSTEM                              0.18                
UPDATE                              0.06                


Query Type                          Total MB Memory     
INSERT_VALUES                       8991.0              
AGGREGATE_JOIN_GROUPBY              7240.000000000001   
UNKNOWN                             6534.0              
DELETE                              6201.0              
SCAN                                2077.0              
AGGREGATE_SCAN_GROUPBY              73.0                
AGGREGATE_JOIN_NO_GROUPBY           48.0                
AGGREGATE_SCAN_NO_GROUPBY           46.0                
JOIN_NO_AGGREGATE                   34.0                
SYSTEM                              1.0                 
UPDATE                              1.0                 


Query Type                          Frequency       Total Time      Total MB Memory
Read                                5782            7587.0          9513.0         
Write                               3591            1525.0          15193.0        


Query Type                          Frequency       Total Time      Total MB Memory
insert_select                       17              5.0             2.0            
insert_values                       3209            1175.0          8991.0         
