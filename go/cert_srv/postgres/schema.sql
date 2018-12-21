CREATE TABLE TRCs (
    IsdID integer NOT NULL,
    Version bigint NOT NULL,
    Data jsonb NOT NULL,
    PRIMARY KEY (IsdID, Version)
);

CREATE TABLE IssuerCerts (
    RowID BIGSERIAL PRIMARY KEY,
    IsdID integer NOT NULL,
    AsID bigint NOT NULL,
    Version bigint NOT NULL,
    Data jsonb NOT NULL,
    UNIQUE (IsdID, AsID, Version)
);

CREATE TABLE LeafCerts (
    IsdID integer NOT NULL,
    AsID bigint NOT NULL,
    Version bigint NOT NULL,
    Data jsonb NOT NULL,
    PRIMARY KEY (IsdID, AsID, Version)
);

CREATE TABLE Chains (
    IsdID integer NOT NULL,
    AsID bigint NOT NULL,
    Version bigint NOT NULL,
    OrderKey integer NOT NULL,
    IssCertsRowID bigint NOT NULL,
    PRIMARY KEY (IsdID, AsID, Version, OrderKey),
    FOREIGN KEY (IssCertsRowID) REFERENCES IssuerCerts(RowID)
);

CREATE TABLE CustKeys (
    IsdID integer NOT NULL,
    AsID bigint NOT NULL,
    Version bigint NOT NULL,
    Key bytea NOT NULL,
    PRIMARY KEY (IsdID, AsID)
);

CREATE TABLE CustKeysLog (
    IsdID integer NOT NULL,
    AsID bigint NOT NULL,
    Version bigint NOT NULL,
    Key bytea NOT NULL,
    PRIMARY KEY (IsdID, AsID, Version)
);
