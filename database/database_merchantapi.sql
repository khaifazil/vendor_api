DROP DATABASE IF EXISTS merchant_api_db;

CREATE DATABASE merchant_api_db;

USE merchant_api_db;

DROP TABLE IF EXISTS merchants;
DROP TABLE IF EXISTS merchant_branches;
DROP TABLE IF EXISTS consumed_vouchers;
DROP TABLE IF EXISTS api_key;

create table merchants
(
    Merchant_ID varchar(45)       not null
        primary key,
    Name        varchar(45)       not null,
    is_active   tinyint default 1 not null,
    constraint Name_UNIQUE
        unique (Name),
    constraint merchant_ID_UNIQUE
        unique (Merchant_ID)
);

create table merchant_branches
(
    Branch_ID      varchar(50)   not null
        primary key,
    Name           varchar(45)   not null,
    Branch_Code    varchar(50)   not null,
    MerchantID     varchar(45)   not null,
    Amount_owed    int default 0 not null,
    Amount_claimed int default 0 not null
);

create table consumed_vouchers
(
    VID          varchar(30)       not null
        primary key,
    Branch_ID    varchar(45)       not null,
    Customer_ID  varchar(45)       not null,
    Amount       int               not null,
    Is_Consumed  tinyint default 0 not null,
    Is_Claimed   tinyint default 0 not null,
    Is_Validated tinyint default 0 not null,
    constraint Voucher_ID_UNIQUE
        unique (VID)
);

create table api_key
(
    ID  int         not null
        primary key auto_increment,
    API_Keys varchar(100) not null,
    constraint API_UNIQUE
        unique (API_Keys)
);

INSERT INTO api_key(API_Keys) VALUE ("b443d88be15b13d50c24a39ae1d3f756be9bcc60d88d55e7e70246c74eb94a6a");