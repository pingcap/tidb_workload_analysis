CREATE INDEX idx_block_number ON ethereum.receipts (block_number);
CREATE INDEX idx_from_address_block_number_token_address ON ethereum.token_transfers (from_address, block_number, token_address);
CREATE INDEX idx_token_address ON ethereum.token_transfers (token_address);
CREATE INDEX idx_block_number ON ethereum.transactions (block_number);
CREATE INDEX idx_from_address ON ethereum.transactions (from_address);
CREATE INDEX idx_hash ON ethereum.transactions (hash);
CREATE INDEX idx_to_address_block_timestamp_value ON ethereum.transactions (to_address, block_timestamp, `value`);