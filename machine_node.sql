CREATE TABLE `machine_node` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'auto increment id',
    `host_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'host name',
    `port` int NOT NULL COMMENT 'port',
    `type` int NOT NULL COMMENT 'node type: 1:ACTUAL or 2:CONTAINER',
    `started_at` datetime NOT NULL COMMENT 'node started time',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='DB WorkerID Assigner for UID Generator';