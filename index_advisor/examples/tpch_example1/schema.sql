CREATE DATABASE tpch;
USE tpch;

CREATE TABLE tpch.`customer` (
                                 `C_CUSTKEY` bigint(20) NOT NULL,
                                 `C_NAME` varchar(25) NOT NULL,
                                 `C_ADDRESS` varchar(40) NOT NULL,
                                 `C_NATIONKEY` bigint(20) NOT NULL,
                                 `C_PHONE` char(15) NOT NULL,
                                 `C_ACCTBAL` decimal(15,2) NOT NULL,
                                 `C_MKTSEGMENT` char(10) NOT NULL,
                                 `C_COMMENT` varchar(117) NOT NULL,
                                 PRIMARY KEY (`C_CUSTKEY`) /*T![clustered_index] CLUSTERED */
);

CREATE TABLE tpch.`lineitem` (
                                 `L_ORDERKEY` bigint(20) NOT NULL,
                                 `L_PARTKEY` bigint(20) NOT NULL,
                                 `L_SUPPKEY` bigint(20) NOT NULL,
                                 `L_LINENUMBER` bigint(20) NOT NULL,
                                 `L_QUANTITY` decimal(15,2) NOT NULL,
                                 `L_EXTENDEDPRICE` decimal(15,2) NOT NULL,
                                 `L_DISCOUNT` decimal(15,2) NOT NULL,
                                 `L_TAX` decimal(15,2) NOT NULL,
                                 `L_RETURNFLAG` char(1) NOT NULL,
                                 `L_LINESTATUS` char(1) NOT NULL,
                                 `L_SHIPDATE` date NOT NULL,
                                 `L_COMMITDATE` date NOT NULL,
                                 `L_RECEIPTDATE` date NOT NULL,
                                 `L_SHIPINSTRUCT` char(25) NOT NULL,
                                 `L_SHIPMODE` char(10) NOT NULL,
                                 `L_COMMENT` varchar(44) NOT NULL,
                                 PRIMARY KEY (`L_ORDERKEY`,`L_LINENUMBER`) /*T![clustered_index] CLUSTERED */
);

CREATE TABLE tpch.`nation` (
                               `N_NATIONKEY` bigint(20) NOT NULL,
                               `N_NAME` char(25) NOT NULL,
                               `N_REGIONKEY` bigint(20) NOT NULL,
                               `N_COMMENT` varchar(152) DEFAULT NULL,
                               PRIMARY KEY (`N_NATIONKEY`) /*T![clustered_index] CLUSTERED */
);

CREATE TABLE tpch.`orders` (
                               `O_ORDERKEY` bigint(20) NOT NULL,
                               `O_CUSTKEY` bigint(20) NOT NULL,
                               `O_ORDERSTATUS` char(1) NOT NULL,
                               `O_TOTALPRICE` decimal(15,2) NOT NULL,
                               `O_ORDERDATE` date NOT NULL,
                               `O_ORDERPRIORITY` char(15) NOT NULL,
                               `O_CLERK` char(15) NOT NULL,
                               `O_SHIPPRIORITY` bigint(20) NOT NULL,
                               `O_COMMENT` varchar(79) NOT NULL,
                               PRIMARY KEY (`O_ORDERKEY`) /*T![clustered_index] CLUSTERED */
);

CREATE TABLE tpch.`part` (
                             `P_PARTKEY` bigint(20) NOT NULL,
                             `P_NAME` varchar(55) NOT NULL,
                             `P_MFGR` char(25) NOT NULL,
                             `P_BRAND` char(10) NOT NULL,
                             `P_TYPE` varchar(25) NOT NULL,
                             `P_SIZE` bigint(20) NOT NULL,
                             `P_CONTAINER` char(10) NOT NULL,
                             `P_RETAILPRICE` decimal(15,2) NOT NULL,
                             `P_COMMENT` varchar(23) NOT NULL,
                             PRIMARY KEY (`P_PARTKEY`) /*T![clustered_index] CLUSTERED */
);


CREATE TABLE tpch.`partsupp` (
                                 `PS_PARTKEY` bigint(20) NOT NULL,
                                 `PS_SUPPKEY` bigint(20) NOT NULL,
                                 `PS_AVAILQTY` bigint(20) NOT NULL,
                                 `PS_SUPPLYCOST` decimal(15,2) NOT NULL,
                                 `PS_COMMENT` varchar(199) NOT NULL,
                                 PRIMARY KEY (`PS_PARTKEY`,`PS_SUPPKEY`) /*T![clustered_index] CLUSTERED */
);

CREATE TABLE tpch.`region` (
                               `R_REGIONKEY` bigint(20) NOT NULL,
                               `R_NAME` char(25) NOT NULL,
                               `R_COMMENT` varchar(152) DEFAULT NULL,
                               PRIMARY KEY (`R_REGIONKEY`) /*T![clustered_index] CLUSTERED */
);

CREATE TABLE tpch.`supplier` (
                                 `S_SUPPKEY` bigint(20) NOT NULL,
                                 `S_NAME` char(25) NOT NULL,
                                 `S_ADDRESS` varchar(40) NOT NULL,
                                 `S_NATIONKEY` bigint(20) NOT NULL,
                                 `S_PHONE` char(15) NOT NULL,
                                 `S_ACCTBAL` decimal(15,2) NOT NULL,
                                 `S_COMMENT` varchar(101) NOT NULL,
                                 PRIMARY KEY (`S_SUPPKEY`) /*T![clustered_index] CLUSTERED */
);
