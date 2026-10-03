CREATE TABLE IF NOT EXISTS part_stock_operations (
    id BIGSERIAL PRIMARY KEY,
    operation_id TEXT NOT NULL,
    part_id BIGINT NOT NULL,
    operation_type TEXT NOT NULL,
    amount INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_part_stock_operation UNIQUE (operation_id, part_id, operation_type)
);

CREATE INDEX IF NOT EXISTS idx_part_stock_operations_op_id ON part_stock_operations (operation_id);
