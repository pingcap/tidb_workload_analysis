create database ethereum;
use ethereum;

CREATE TABLE `blocks` (
                          `timestamp` bigint(20) DEFAULT NULL,
                          `number` bigint(20) DEFAULT NULL,
                          `hash` varchar(66) DEFAULT NULL,
                          `parent_hash` varchar(66) DEFAULT NULL,
                          `nonce` varchar(42) DEFAULT NULL,
                          `sha3_uncles` varchar(66) DEFAULT NULL,
                          `logs_bloom` text DEFAULT NULL,
                          `transactions_root` varchar(66) DEFAULT NULL,
                          `state_root` varchar(66) DEFAULT NULL,
                          `receipts_root` varchar(66) DEFAULT NULL,
                          `miner` varchar(42) DEFAULT NULL,
                          `difficulty` decimal(38,0) DEFAULT NULL,
                          `total_difficulty` decimal(38,0) DEFAULT NULL,
                          `size` bigint(20) DEFAULT NULL,
                          `extra_data` text DEFAULT NULL,
                          `gas_limit` bigint(20) DEFAULT NULL,
                          `gas_used` bigint(20) DEFAULT NULL,
                          `transaction_count` bigint(20) DEFAULT NULL,
                          `base_fee_per_gas` bigint(20) DEFAULT NULL
);

CREATE TABLE `transactions` (
                                `hash` varchar(66) DEFAULT NULL,
                                `nonce` bigint(20) DEFAULT NULL,
                                `block_hash` varchar(66) DEFAULT NULL,
                                `block_number` bigint(20) DEFAULT NULL,
                                `transaction_index` bigint(20) DEFAULT NULL,
                                `from_address` varchar(42) DEFAULT NULL,
                                `to_address` varchar(42) DEFAULT NULL,
                                `value` decimal(38,0) DEFAULT NULL,
                                `gas` bigint(20) DEFAULT NULL,
                                `gas_price` bigint(20) DEFAULT NULL,
                                `input` text DEFAULT NULL,
                                `block_timestamp` bigint(20) DEFAULT NULL,
                                `max_fee_per_gas` bigint(20) DEFAULT NULL,
                                `max_priority_fee_per_gas` bigint(20) DEFAULT NULL,
                                `transaction_type` bigint(20) DEFAULT NULL
);


CREATE TABLE `token_transfers` (
                                   `token_address` varchar(42) DEFAULT NULL,
                                   `from_address` varchar(42) DEFAULT NULL,
                                   `to_address` varchar(42) DEFAULT NULL,
                                   `value` varchar(78) DEFAULT NULL COMMENT 'Postgresql use numeric(78), while the max_value of Decimal is decimal(65), thus use string here',
                                   `transaction_hash` varchar(66) DEFAULT NULL,
                                   `log_index` bigint(20) DEFAULT NULL,
                                   `block_number` bigint(20) DEFAULT NULL
);


CREATE TABLE `receipts` (
                            `transaction_hash` varchar(66) DEFAULT NULL,
                            `transaction_index` bigint(20) DEFAULT NULL,
                            `block_hash` varchar(66) DEFAULT NULL,
                            `block_number` bigint(20) DEFAULT NULL,
                            `cumulative_gas_used` bigint(20) DEFAULT NULL,
                            `gas_used` bigint(20) DEFAULT NULL,
                            `contract_address` varchar(42) DEFAULT NULL,
                            `root` varchar(66) DEFAULT NULL,
                            `status` bigint(20) DEFAULT NULL,
                            `effective_gas_price` bigint(20) DEFAULT NULL
);


CREATE TABLE `logs` (
                        `log_index` bigint(20) DEFAULT NULL,
                        `transaction_hash` varchar(66) DEFAULT NULL,
                        `transaction_index` bigint(20) DEFAULT NULL,
                        `block_hash` varchar(66) DEFAULT NULL,
                        `block_number` bigint(20) DEFAULT NULL,
                        `address` varchar(42) DEFAULT NULL,
                        `data` text DEFAULT NULL,
                        `topics` text DEFAULT NULL
);

CREATE TABLE `contracts` (
                             `address` char(42) DEFAULT NULL,
                             `bytecode` text DEFAULT NULL,
                             `function_sighashes` text DEFAULT NULL,
                             `is_erc20` tinyint(1) DEFAULT NULL,
                             `is_erc721` tinyint(1) DEFAULT NULL,
                             `block_number` bigint(20) DEFAULT NULL
);

CREATE TABLE `tokens` (
                          `address` char(42) DEFAULT NULL,
                          `symbol` text DEFAULT NULL,
                          `name` text DEFAULT NULL,
                          `decimals` bigint(20) DEFAULT NULL,
                          `total_supply` decimal(38,0) DEFAULT NULL,
                          `block_number` bigint(20) DEFAULT NULL
);

CREATE TABLE `traces` (
                          `block_number` bigint(20) DEFAULT NULL,
                          `transaction_hash` varchar(66) DEFAULT NULL,
                          `transaction_index` bigint(20) DEFAULT NULL,
                          `from_address` varchar(42) DEFAULT NULL,
                          `to_address` varchar(42) DEFAULT NULL,
                          `value` decimal(38,0) DEFAULT NULL,
                          `input` text DEFAULT NULL,
                          `output` text DEFAULT NULL,
                          `trace_type` varchar(16) DEFAULT NULL,
                          `call_type` varchar(16) DEFAULT NULL,
                          `reward_type` varchar(16) DEFAULT NULL,
                          `gas` bigint(20) DEFAULT NULL,
                          `gas_used` bigint(20) DEFAULT NULL,
                          `subtraces` bigint(20) DEFAULT NULL,
                          `trace_address` text DEFAULT NULL,
                          `error` text DEFAULT NULL,
                          `status` bigint(20) DEFAULT NULL,
                          `trace_id` text DEFAULT NULL
);