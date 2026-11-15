CREATE TABLE IF NOT EXISTS part_stock_operations (
    id BIGSERIAL PRIMARY KEY,
    part_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    amount INT NOT NULL,
    operation_id TEXT NOT NULL,
    operation_type TEXT NOT NULL,
    CONSTRAINT uq_part_stock_operation UNIQUE (operation_id, part_id, operation_type)
);

CREATE INDEX IF NOT EXISTS idx_part_stock_operations_op_id ON part_stock_operations (operation_id);
