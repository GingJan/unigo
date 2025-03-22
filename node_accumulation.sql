create table node_accumulation (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'auto increment id',
    host_name   VARCHAR(64) NOT NULL COMMENT 'host name',
    port        VARCHAR(64) NOT NULL COMMENT 'port',
    type        int         NOT NULL COMMENT 'node type: ACTUAL or CONTAINER',
    started_at  datetime    NOT NULL COMMENT 'node started time',
) COMMENT 'DB WorkerID Assigner for UID Generator';