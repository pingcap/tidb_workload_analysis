import sqlite3 as lite
import argparse
import os
import sys
from urllib.request import pathname2url
import ctypes
import json
import re
import math

lib = ctypes.CDLL('./analyze.so')
lib.analyze.restype = ctypes.c_char_p

"""
Mainly calls analyze.so library to get list of markers for a SQL query.
Markers capture logical operators in query. See analyze.go for more details.
"""
def analyze_one_query(sql):
        #few hacks to clean up SQL
        sql = sql.replace('\n',' ')
        sql = sql.strip()
        key_string = (lib.analyze(sql.encode("utf-8"), True))
        key_string = json.loads(key_string.decode("utf-8"))
        if (len(key_string) == 0):
            sql = sql.lower()
            # Analyze the query text if the parser fails.
            if re.search('.*insert.*values.*',sql):
                key_string = ['Insert']
            elif re.search('^analyze.*',sql):
                key_string = ['Analyze']
            else:
                key_string = ['None']
        key_string = str(key_string).replace("'","")
        return key_string

# Total number of joins in a SQL using markers.
def number_of_joins(markers):
        num = markers.count('InnerJoin')+markers.count('LeftJoin')+markers.count('RightJoin')
        return num

# Total number of aggregate expressions in a SQL using markers.
def number_of_aggregate(markers):
        return markers.count('Aggregate')

# Total number of GROUP BY operators in a SQL using markers.
def number_of_groupby(markers):
        return markers.count('GroupBy')+markers.count('Distinct')

"""
Compute query type using the list of markers for a query. The list of query types is an initial one and
more types can be added based on the markers/operators in the SQL.
"""
def query_type(markers):
        if (markers == '[Insert]'):
            return 'INSERT_VALUES'
        elif (markers.count('Delete') == 1 and markers.count('Select') == 0):
            return 'DELETE'
        elif (markers.count('Update') == 1 and markers.count('Select') == 0):
            return 'UPDATE'
        elif (markers == '[Analyze]'):
            return 'ANALYZE'
        elif (markers == '[None]'):
            return 'UNKNOWN'
        elif (markers.count('Explain') == 1):
            return 'EXPLAIN'
        elif (markers.count('System') == 1):
            return 'SYSTEM'
        else:
            num_agg = number_of_aggregate(markers)
            num_groupby = number_of_groupby(markers)
            num_join = number_of_joins(markers)
            if (num_agg == 0 and num_groupby == 0 and num_join == 0):
                return 'SCAN'
            elif (num_agg > 0 and num_groupby == 0 and num_join == 0):
                return 'AGGREGATE_SCAN_NO_GROUPBY'
            elif (num_agg > 0 and num_groupby == 0 and num_join > 0):
                return 'AGGREGATE_JOIN_NO_GROUPBY'
            elif (num_groupby > 0 and num_join == 0):
                return 'AGGREGATE_SCAN_GROUPBY'
            elif (num_groupby > 0 and num_join > 0):
                return 'AGGREGATE_JOIN_GROUPBY'
            elif (num_groupby == 0 and num_agg == 0 and num_join > 0):
                return 'JOIN_NO_AGGREGATE'
            else:
                return 'FIX-IT'

"""
Find total resources consumed by queries broken down by query_type.
resource can be either query time or memory used bt the query.
"""
def report_by_query_resource(cur, time_or_mem):
        column1 = 'Query Type'
        if time_or_mem == 'time':
            column2 = 'Total Query Time in Seconds'
            report_by_query_resource_cur = cur.execute("""
                select query_type, sum(total_query_time) from unique_queries group by query_type order by sum(total_query_time) desc
            """)
        elif time_or_mem == 'memory':
            column2 = 'Total MB Memory'
            report_by_query_resource_cur = cur.execute("""
                select query_type, sum(total_mem) from unique_queries group by query_type order by sum(total_mem) desc
            """)
        else:
            print("\n invalid choice in report_by_query_resource ")
            return
        one_row = report_by_query_resource_cur.fetchone() 
        if one_row is not None:
            print("\n")
            print(f'{column1:<35}', f'{column2:<20}')
        while one_row is not None:
            total_query_resource = math.ceil(one_row[1]*100)/100
            if time_or_mem == 'memory':
                total_query_resource = math.ceil(total_query_resource/1024.00/1024.00)/100*100
            print(f'{one_row[0]:<35}', f'{total_query_resource:<20}')
            one_row = report_by_query_resource_cur.fetchone() 

"""
Similar to report_by_query_resource, report_by_frequency finds the total frequency of
unique queries broken down by query_type.
"""
def report_by_frequency(cur):
        report_by_frequency_cur = cur.execute("""
            select query_type, sum(frequency) from unique_queries group by query_type order by sum(frequency) desc
        """)
        one_row = report_by_frequency_cur.fetchone() 
        if one_row is not None:
            print("\n")
            column1 = 'Query Type'
            column2 = 'Frequency'
            print(f'{column1:<35}', f'{column2:<10}')
        while one_row is not None:
            print(f'{one_row[0]:<35}', f'{one_row[1]:<10}')
            one_row = report_by_frequency_cur.fetchone() 

"""
Find total compute time and memory of read and write queries.
Write queries are those with types: Insert, Delete or Update.
All other queries are read queries.
"""
def read_vs_write_report(cur):
        report_cur = cur.execute("""
            select query_markers, frequency, total_query_time, total_mem  from unique_queries
        """)
        one_row = report_cur.fetchone() 
        if one_row is not None:
            print("\n")
            column1 = 'Query Type'
            column2 = 'Frequency'
            column3 = 'Total Time'
            column4 = 'Total MB Memory'
            write = 'Write'
            read = 'Read'
            read_freq = 0
            read_total_time = 0.0
            read_total_mem = 0.0
            write_freq = 0
            write_total_time = 0.0
            write_total_mem = 0.0
            print(f'{column1:<35}', f'{column2:<15}', f'{column3:<15}', f'{column4:<15}')
        else:
            return
        while one_row is not None:
            markers = one_row[0]
            if (markers.count('Delete') > 0 or markers.count('Insert') > 0 or markers.count('Update') > 0):
                write_freq = write_freq+one_row[1]
                write_total_time = write_total_time+one_row[2]
                write_total_mem = write_total_mem+one_row[3]
            elif (markers.count('Select') > 0):
                read_freq = read_freq+one_row[1]
                read_total_time = read_total_time+one_row[2]
                read_total_mem = read_total_mem+one_row[3]
            one_row = report_cur.fetchone() 
        write_total_time = math.ceil(write_total_time)*100/100
        write_total_mem = math.ceil(write_total_mem/1024.0/1024.0)*100/100
        read_total_time = math.ceil(read_total_time)*100/100
        read_total_mem = math.ceil(read_total_mem/1024.0/1024.0)*100/100
        print(f'{read:<35}', f'{read_freq:<15}', f'{read_total_time:<15}', f'{read_total_mem:<15}')
        print(f'{write:<35}', f'{write_freq:<15}', f'{write_total_time:<15}', f'{write_total_mem:<15}')

"""
Compare metrics for insert values vs insert select. 
Metrics include frequency, compute time and memory.
"""
def insert_select_vs_insert_values_report(cur):
        report_cur = cur.execute("""
            select query_type, frequency, total_query_time, total_mem  
            from unique_queries
            where query_markers like '%Insert%'
        """)
        one_row = report_cur.fetchone() 
        if one_row is not None:
            print("\n")
            column1 = 'Query Type'
            column2 = 'Frequency'
            column3 = 'Total Time'
            column4 = 'Total MB Memory'
            insert_values = 'insert_values'
            insert_select = 'insert_select'
            insert_select_freq = 0
            insert_select_total_time = 0.0
            insert_select_total_mem = 0.0
            insert_values_freq = 0
            insert_values_total_time = 0.0
            insert_values_total_mem = 0.0
            print(f'{column1:<35}', f'{column2:<15}', f'{column3:<15}', f'{column4:<15}')
        else:
            return
        while one_row is not None:
            if (one_row[0] == 'INSERT_VALUES'):
                insert_values_freq = insert_values_freq+one_row[1]
                insert_values_total_time = insert_values_total_time+one_row[2]
                insert_values_total_mem = insert_values_total_mem+one_row[3]
            else:
                insert_select_freq = insert_select_freq+one_row[1]
                insert_select_total_time = insert_select_total_time+one_row[2]
                insert_select_total_mem = insert_select_total_mem+one_row[3]
            one_row = report_cur.fetchone() 
        insert_values_total_time = math.ceil(insert_values_total_time)*100/100
        insert_values_total_mem = math.ceil(insert_values_total_mem/1024.0/1024.0)*100/100
        insert_select_total_time = math.ceil(insert_select_total_time)*100/100
        insert_select_total_mem = math.ceil(insert_select_total_mem/1024.0/1024.0)*100/100
        print(f'{insert_select:<35}', f'{insert_select_freq:<15}', f'{insert_select_total_time:<15}', f'{insert_select_total_mem:<15}')
        print(f'{insert_values:<35}', f'{insert_values_freq:<15}', f'{insert_values_total_time:<15}', f'{insert_values_total_mem:<15}')
"""
This is main routine for analysis. It first finds unique queries and stores them into a new table called unique_queroes.
For each unqiue query, we compute total frequency, total compute and query type. The function also produces 5 reports:
 - Frequency analysis by query type.
 - Total compute time by query type.
 - Total memory consumption by query type.
 - Frequency, compute time and memory of read vs write requests.
 - Frequency, compute time and memory of insert values vs insert select.
insert_select_vs_insert_values_report(cur)
"""
def analyze_queries(con):
        con.create_function("analyzeOneQuery", 1, analyze_one_query)
        con.create_function("numberOfJoins", 1, number_of_joins)
        con.create_function("queryType", 1, query_type)
        cur = con.cursor()
        cur.execute("drop table if exists unique_queries")
        cur.execute("""
            create table unique_queries(
                digest text,
                sql_statement text,
                frequency decimal,
                total_query_time decimal,
                max_query_time decimal,
                min_query_time decimal,
                total_mem decimal,
                max_mem decimal,
                min_mem decimal,
                query_markers text,
                number_of_joins tinyint,
                query_type text)
        """)
        insert_statement = """
            INSERT INTO unique_queries(
                digest, sql_statement, frequency,
                total_query_time, max_query_time, min_query_time,
                total_mem, max_mem, min_mem,
                query_markers, number_of_joins, query_type
            )
            SELECT digest, max(sql_statement), count(*) as frequency,
                   sum(CAST(query_time as decimal)) as total_query_time,
                   max(CAST(query_time as decimal)) as max_query_time,
                   min(CAST(query_time as decimal)) as min_query_time,
                   sum(CAST(mem_max as decimal)) as total_mem,
                   max(CAST(mem_max as decimal)) as max_mem,
                   min(CAST(mem_max as decimal)) as min_mem,
                   analyzeOneQuery(sql_statement),
                   numberOfJoins(analyzeOneQuery(sql_statement)),
                   queryType(analyzeOneQuery(sql_statement))
            FROM slowlog
            GROUP BY digest 
            order by max_query_time desc
        """
        cur.execute(insert_statement)
        report_by_frequency(cur)
        report_by_query_resource(cur,'time')
        report_by_query_resource(cur,'memory')
        read_vs_write_report(cur)
        insert_select_vs_insert_values_report(cur)

"""
Main API for the tool. Check if the test database is sqlite format.
Also, check if the sqlite database has a table called slowlog.
"""
def main():
    parser = argparse.ArgumentParser(description='Workload analysis.')
    parser.add_argument('--test_database',  help='sqllite name with log data')
    args = parser.parse_args()
    test_database = args.test_database

    if test_database != "":
        try:
            dburi = 'file:{}?mode=rw'.format(pathname2url(test_database))
            con = lite.connect(dburi, uri=True)
        except lite.OperationalError:
            print("\n incorrect database name:",test_database)
            sys.exit()
        cur = con.cursor()
        listOfTables = cur.execute("""SELECT name FROM sqlite_master WHERE type='table' AND name='slowlog'; """).fetchall()
        if listOfTables == []:
            print('slowlog table is missing')
        else:
            analyze_queries(con)
            con.commit()
        con.close()
    else:
        print ("usage: python3 query_analysis.py --test_database database_name")

if __name__ == '__main__':
    main()
