-- ptt-monkat : MariaDB master data
-- One row per (device, OID). The poller is device-agnostic: everything it needs
-- to reach a device and interpret the raw SNMP value lives in this table.

CREATE DATABASE IF NOT EXISTS ptt_monkat
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ptt_monkat;

CREATE TABLE IF NOT EXISTS snmp_targets (
  id                BIGINT UNSIGNED    NOT NULL AUTO_INCREMENT,
  device_name       VARCHAR(128)       NOT NULL,
  ip_address        VARCHAR(45)        NOT NULL,              -- fits an IPv6 literal
  oid               VARCHAR(255)       NOT NULL,              -- e.g. 1.3.6.1.4.1.9.9.109.1.1.1.1.6.1
  metric_name       VARCHAR(64)        NOT NULL,              -- CPU, Temp, RxPower, Load, ...
  unit              VARCHAR(32)        NOT NULL DEFAULT '',   -- %, C, dBm, V, bps
  multiplier        DECIMAL(18,8)      NOT NULL DEFAULT 1.00000000,
  is_active         TINYINT(1)         NOT NULL DEFAULT 1,

  snmp_community    VARCHAR(64)        NOT NULL DEFAULT 'public',
  snmp_version      ENUM('1','2c','3') NOT NULL DEFAULT '2c',
  snmp_port         SMALLINT UNSIGNED  NOT NULL DEFAULT 161,
  poll_interval_sec INT UNSIGNED       NOT NULL DEFAULT 60,

  created_at        TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP
                                       ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  UNIQUE KEY uq_target (ip_address, oid),
  KEY idx_poller (is_active, ip_address)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- multiplier is DECIMAL, not DOUBLE: scaling factors like 0.1 must be exact.
-- A dBm reading off by a float epsilon becomes a support ticket.

-- Seed rows for stage 2 to poll. Point them at something real before enabling.
-- INSERT INTO snmp_targets (device_name, ip_address, oid, metric_name, unit, multiplier, is_active) VALUES
--   ('core-rtr-01', '10.0.0.1', '1.3.6.1.4.1.9.9.109.1.1.1.1.6.1', 'CPU',  '%',   1.0, 0),
--   ('core-rtr-01', '10.0.0.1', '1.3.6.1.4.1.9.9.13.1.3.1.3.1',    'Temp', 'C',   1.0, 0),
--   ('dwdm-01',     '10.0.0.5', '1.3.6.1.4.1.1000.1.1.1.2.1',      'RxPower', 'dBm', 0.1, 0);
