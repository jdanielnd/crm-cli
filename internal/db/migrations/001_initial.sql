-- up

-- People
CREATE TABLE IF NOT EXISTS people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT,
    email TEXT,
    phone TEXT,
    title TEXT,
    company TEXT,
    location TEXT,
    notes TEXT,
    summary TEXT,
    org_id INTEGER REFERENCES organizations(id),
    archived INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Organizations
CREATE TABLE IF NOT EXISTS organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    domain TEXT,
    industry TEXT,
    notes TEXT,
    summary TEXT,
    archived INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Interactions
CREATE TABLE IF NOT EXISTS interactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL CHECK(type IN ('call', 'email', 'meeting', 'note', 'message')),
    subject TEXT,
    content TEXT,
    direction TEXT CHECK(direction IN ('inbound', 'outbound')),
    occurred_at TEXT NOT NULL DEFAULT (datetime('now')),
    archived INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Junction: interactions <-> people
CREATE TABLE IF NOT EXISTS interaction_people (
    interaction_id INTEGER NOT NULL REFERENCES interactions(id) ON DELETE CASCADE,
    person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    PRIMARY KEY (interaction_id, person_id)
);

-- Deals
CREATE TABLE IF NOT EXISTS deals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    value REAL,
    stage TEXT NOT NULL DEFAULT 'lead' CHECK(stage IN ('lead', 'prospect', 'proposal', 'negotiation', 'won', 'lost')),
    person_id INTEGER REFERENCES people(id),
    org_id INTEGER REFERENCES organizations(id),
    notes TEXT,
    closed_at TEXT,
    archived INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Tasks
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT,
    person_id INTEGER REFERENCES people(id),
    deal_id INTEGER REFERENCES deals(id),
    due_at TEXT,
    priority TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low', 'medium', 'high')),
    completed INTEGER NOT NULL DEFAULT 0,
    completed_at TEXT,
    archived INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Tags
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- Polymorphic tagging
CREATE TABLE IF NOT EXISTS taggings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL CHECK(entity_type IN ('person', 'organization', 'deal', 'interaction')),
    entity_id INTEGER NOT NULL,
    UNIQUE(tag_id, entity_type, entity_id)
);

-- Custom fields (key/value on any entity)
CREATE TABLE IF NOT EXISTS custom_fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    field_name TEXT NOT NULL,
    field_value TEXT NOT NULL,
    UNIQUE(entity_type, entity_id, field_name)
);

-- Relationships (person-to-person)
CREATE TABLE IF NOT EXISTS relationships (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    related_person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK(type IN ('colleague', 'friend', 'manager', 'mentor', 'referred-by')),
    notes TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    CHECK(person_id != related_person_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_people_org ON people(org_id) WHERE archived = 0;
CREATE INDEX IF NOT EXISTS idx_people_email ON people(email) WHERE archived = 0;
CREATE INDEX IF NOT EXISTS idx_interaction_people_person ON interaction_people(person_id);
CREATE INDEX IF NOT EXISTS idx_deals_person ON deals(person_id) WHERE archived = 0;
CREATE INDEX IF NOT EXISTS idx_deals_org ON deals(org_id) WHERE archived = 0;
CREATE INDEX IF NOT EXISTS idx_deals_stage ON deals(stage) WHERE archived = 0;
CREATE INDEX IF NOT EXISTS idx_tasks_person ON tasks(person_id) WHERE archived = 0;
CREATE INDEX IF NOT EXISTS idx_tasks_deal ON tasks(deal_id) WHERE archived = 0;
CREATE INDEX IF NOT EXISTS idx_tasks_due ON tasks(due_at) WHERE archived = 0 AND completed = 0;
CREATE INDEX IF NOT EXISTS idx_taggings_entity ON taggings(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_custom_fields_entity ON custom_fields(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_relationships_person ON relationships(person_id);
CREATE INDEX IF NOT EXISTS idx_relationships_related ON relationships(related_person_id);

-- FTS5: people
CREATE VIRTUAL TABLE IF NOT EXISTS people_fts USING fts5(
    first_name, last_name, email, company, notes,
    content='people', content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS people_ai AFTER INSERT ON people BEGIN
    INSERT INTO people_fts(rowid, first_name, last_name, email, company, notes)
    VALUES (new.id, new.first_name, new.last_name, new.email, new.company, new.notes);
END;

CREATE TRIGGER IF NOT EXISTS people_ad AFTER DELETE ON people BEGIN
    INSERT INTO people_fts(people_fts, rowid, first_name, last_name, email, company, notes)
    VALUES ('delete', old.id, old.first_name, old.last_name, old.email, old.company, old.notes);
END;

CREATE TRIGGER IF NOT EXISTS people_au AFTER UPDATE ON people BEGIN
    INSERT INTO people_fts(people_fts, rowid, first_name, last_name, email, company, notes)
    VALUES ('delete', old.id, old.first_name, old.last_name, old.email, old.company, old.notes);
    INSERT INTO people_fts(rowid, first_name, last_name, email, company, notes)
    VALUES (new.id, new.first_name, new.last_name, new.email, new.company, new.notes);
END;

-- FTS5: organizations
CREATE VIRTUAL TABLE IF NOT EXISTS organizations_fts USING fts5(
    name, domain, industry, notes,
    content='organizations', content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS orgs_ai AFTER INSERT ON organizations BEGIN
    INSERT INTO organizations_fts(rowid, name, domain, industry, notes)
    VALUES (new.id, new.name, new.domain, new.industry, new.notes);
END;

CREATE TRIGGER IF NOT EXISTS orgs_ad AFTER DELETE ON organizations BEGIN
    INSERT INTO organizations_fts(organizations_fts, rowid, name, domain, industry, notes)
    VALUES ('delete', old.id, old.name, old.domain, old.industry, old.notes);
END;

CREATE TRIGGER IF NOT EXISTS orgs_au AFTER UPDATE ON organizations BEGIN
    INSERT INTO organizations_fts(organizations_fts, rowid, name, domain, industry, notes)
    VALUES ('delete', old.id, old.name, old.domain, old.industry, old.notes);
    INSERT INTO organizations_fts(rowid, name, domain, industry, notes)
    VALUES (new.id, new.name, new.domain, new.industry, new.notes);
END;

-- FTS5: interactions
CREATE VIRTUAL TABLE IF NOT EXISTS interactions_fts USING fts5(
    subject, content,
    content='interactions', content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS interactions_ai AFTER INSERT ON interactions BEGIN
    INSERT INTO interactions_fts(rowid, subject, content)
    VALUES (new.id, new.subject, new.content);
END;

CREATE TRIGGER IF NOT EXISTS interactions_ad AFTER DELETE ON interactions BEGIN
    INSERT INTO interactions_fts(interactions_fts, rowid, subject, content)
    VALUES ('delete', old.id, old.subject, old.content);
END;

CREATE TRIGGER IF NOT EXISTS interactions_au AFTER UPDATE ON interactions BEGIN
    INSERT INTO interactions_fts(interactions_fts, rowid, subject, content)
    VALUES ('delete', old.id, old.subject, old.content);
    INSERT INTO interactions_fts(rowid, subject, content)
    VALUES (new.id, new.subject, new.content);
END;

-- FTS5: deals
CREATE VIRTUAL TABLE IF NOT EXISTS deals_fts USING fts5(
    title, notes,
    content='deals', content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS deals_ai AFTER INSERT ON deals BEGIN
    INSERT INTO deals_fts(rowid, title, notes)
    VALUES (new.id, new.title, new.notes);
END;

CREATE TRIGGER IF NOT EXISTS deals_ad AFTER DELETE ON deals BEGIN
    INSERT INTO deals_fts(deals_fts, rowid, title, notes)
    VALUES ('delete', old.id, old.title, old.notes);
END;

CREATE TRIGGER IF NOT EXISTS deals_au AFTER UPDATE ON deals BEGIN
    INSERT INTO deals_fts(deals_fts, rowid, title, notes)
    VALUES ('delete', old.id, old.title, old.notes);
    INSERT INTO deals_fts(rowid, title, notes)
    VALUES (new.id, new.title, new.notes);
END;

-- down

DROP TRIGGER IF EXISTS deals_au;
DROP TRIGGER IF EXISTS deals_ad;
DROP TRIGGER IF EXISTS deals_ai;
DROP TABLE IF EXISTS deals_fts;
DROP TRIGGER IF EXISTS interactions_au;
DROP TRIGGER IF EXISTS interactions_ad;
DROP TRIGGER IF EXISTS interactions_ai;
DROP TABLE IF EXISTS interactions_fts;
DROP TRIGGER IF EXISTS orgs_au;
DROP TRIGGER IF EXISTS orgs_ad;
DROP TRIGGER IF EXISTS orgs_ai;
DROP TABLE IF EXISTS organizations_fts;
DROP TRIGGER IF EXISTS people_au;
DROP TRIGGER IF EXISTS people_ad;
DROP TRIGGER IF EXISTS people_ai;
DROP TABLE IF EXISTS people_fts;
DROP TABLE IF EXISTS relationships;
DROP TABLE IF EXISTS custom_fields;
DROP TABLE IF EXISTS taggings;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS deals;
DROP TABLE IF EXISTS interaction_people;
DROP TABLE IF EXISTS interactions;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS people;
