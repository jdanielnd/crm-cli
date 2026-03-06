-- up

-- Organizations
CREATE TABLE organizations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6)))),
  name TEXT NOT NULL,
  domain TEXT,
  industry TEXT,
  location TEXT,
  notes TEXT,
  summary TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- People
CREATE TABLE people (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6)))),
  first_name TEXT NOT NULL,
  last_name TEXT,
  email TEXT,
  phone TEXT,
  title TEXT,
  company TEXT,
  location TEXT,
  linkedin TEXT,
  twitter TEXT,
  website TEXT,
  notes TEXT,
  summary TEXT,
  org_id INTEGER REFERENCES organizations(id),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- Interactions
CREATE TABLE interactions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6)))),
  type TEXT NOT NULL CHECK (type IN ('call', 'email', 'meeting', 'note', 'message')),
  subject TEXT,
  content TEXT,
  direction TEXT CHECK (direction IN ('inbound', 'outbound')),
  occurred_at TEXT NOT NULL DEFAULT (datetime('now')),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- Interaction-person junction (many-to-many)
CREATE TABLE interaction_people (
  interaction_id INTEGER NOT NULL REFERENCES interactions(id),
  person_id INTEGER NOT NULL REFERENCES people(id),
  PRIMARY KEY (interaction_id, person_id)
);

-- Deals
CREATE TABLE deals (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6)))),
  title TEXT NOT NULL,
  value REAL,
  currency TEXT DEFAULT 'USD',
  stage TEXT NOT NULL DEFAULT 'lead' CHECK (stage IN ('lead', 'prospect', 'proposal', 'negotiation', 'won', 'lost')),
  person_id INTEGER REFERENCES people(id),
  org_id INTEGER REFERENCES organizations(id),
  closed_at TEXT,
  notes TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- Tasks
CREATE TABLE tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6)))),
  title TEXT NOT NULL,
  description TEXT,
  due_at TEXT,
  priority TEXT NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
  completed INTEGER NOT NULL DEFAULT 0,
  completed_at TEXT,
  person_id INTEGER REFERENCES people(id),
  deal_id INTEGER REFERENCES deals(id),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- Tags
CREATE TABLE tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE
);

-- Taggings (polymorphic)
CREATE TABLE taggings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tag_id INTEGER NOT NULL REFERENCES tags(id),
  entity_type TEXT NOT NULL CHECK (entity_type IN ('person', 'organization', 'deal', 'interaction')),
  entity_id INTEGER NOT NULL,
  UNIQUE(tag_id, entity_type, entity_id)
);

-- Custom fields (polymorphic)
CREATE TABLE custom_fields (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  entity_type TEXT NOT NULL,
  entity_id INTEGER NOT NULL,
  field_name TEXT NOT NULL,
  field_value TEXT,
  UNIQUE(entity_type, entity_id, field_name)
);

-- Relationships (person-to-person)
CREATE TABLE relationships (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  person_id INTEGER NOT NULL REFERENCES people(id),
  related_person_id INTEGER NOT NULL REFERENCES people(id),
  type TEXT NOT NULL CHECK (type IN ('colleague', 'friend', 'manager', 'report', 'mentor', 'mentee', 'referred-by', 'referred')),
  notes TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(person_id, related_person_id, type)
);

-- FTS5 indexes
CREATE VIRTUAL TABLE people_fts USING fts5(
  first_name, last_name, email, company, notes, summary,
  content=people, content_rowid=id
);

CREATE VIRTUAL TABLE organizations_fts USING fts5(
  name, domain, industry, notes, summary,
  content=organizations, content_rowid=id
);

CREATE VIRTUAL TABLE interactions_fts USING fts5(
  subject, content,
  content=interactions, content_rowid=id
);

CREATE VIRTUAL TABLE deals_fts USING fts5(
  title, notes,
  content=deals, content_rowid=id
);

-- FTS triggers: people
CREATE TRIGGER people_fts_insert AFTER INSERT ON people BEGIN
  INSERT INTO people_fts(rowid, first_name, last_name, email, company, notes, summary)
  VALUES (new.id, new.first_name, new.last_name, new.email, new.company, new.notes, new.summary);
END;

CREATE TRIGGER people_fts_delete AFTER DELETE ON people BEGIN
  INSERT INTO people_fts(people_fts, rowid, first_name, last_name, email, company, notes, summary)
  VALUES ('delete', old.id, old.first_name, old.last_name, old.email, old.company, old.notes, old.summary);
END;

CREATE TRIGGER people_fts_update AFTER UPDATE ON people BEGIN
  INSERT INTO people_fts(people_fts, rowid, first_name, last_name, email, company, notes, summary)
  VALUES ('delete', old.id, old.first_name, old.last_name, old.email, old.company, old.notes, old.summary);
  INSERT INTO people_fts(rowid, first_name, last_name, email, company, notes, summary)
  VALUES (new.id, new.first_name, new.last_name, new.email, new.company, new.notes, new.summary);
END;

-- FTS triggers: organizations
CREATE TRIGGER organizations_fts_insert AFTER INSERT ON organizations BEGIN
  INSERT INTO organizations_fts(rowid, name, domain, industry, notes, summary)
  VALUES (new.id, new.name, new.domain, new.industry, new.notes, new.summary);
END;

CREATE TRIGGER organizations_fts_delete AFTER DELETE ON organizations BEGIN
  INSERT INTO organizations_fts(organizations_fts, rowid, name, domain, industry, notes, summary)
  VALUES ('delete', old.id, old.name, old.domain, old.industry, old.notes, old.summary);
END;

CREATE TRIGGER organizations_fts_update AFTER UPDATE ON organizations BEGIN
  INSERT INTO organizations_fts(organizations_fts, rowid, name, domain, industry, notes, summary)
  VALUES ('delete', old.id, old.name, old.domain, old.industry, old.notes, old.summary);
  INSERT INTO organizations_fts(rowid, name, domain, industry, notes, summary)
  VALUES (new.id, new.name, new.domain, new.industry, new.notes, new.summary);
END;

-- FTS triggers: interactions
CREATE TRIGGER interactions_fts_insert AFTER INSERT ON interactions BEGIN
  INSERT INTO interactions_fts(rowid, subject, content)
  VALUES (new.id, new.subject, new.content);
END;

CREATE TRIGGER interactions_fts_delete AFTER DELETE ON interactions BEGIN
  INSERT INTO interactions_fts(interactions_fts, rowid, subject, content)
  VALUES ('delete', old.id, old.subject, old.content);
END;

CREATE TRIGGER interactions_fts_update AFTER UPDATE ON interactions BEGIN
  INSERT INTO interactions_fts(interactions_fts, rowid, subject, content)
  VALUES ('delete', old.id, old.subject, old.content);
  INSERT INTO interactions_fts(rowid, subject, content)
  VALUES (new.id, new.subject, new.content);
END;

-- FTS triggers: deals
CREATE TRIGGER deals_fts_insert AFTER INSERT ON deals BEGIN
  INSERT INTO deals_fts(rowid, title, notes)
  VALUES (new.id, new.title, new.notes);
END;

CREATE TRIGGER deals_fts_delete AFTER DELETE ON deals BEGIN
  INSERT INTO deals_fts(deals_fts, rowid, title, notes)
  VALUES ('delete', old.id, old.title, old.notes);
END;

CREATE TRIGGER deals_fts_update AFTER UPDATE ON deals BEGIN
  INSERT INTO deals_fts(deals_fts, rowid, title, notes)
  VALUES ('delete', old.id, old.title, old.notes);
  INSERT INTO deals_fts(rowid, title, notes)
  VALUES (new.id, new.title, new.notes);
END;

-- Indexes
CREATE INDEX idx_people_org_id ON people(org_id) WHERE archived = 0;
CREATE INDEX idx_people_email ON people(email) WHERE archived = 0;
CREATE INDEX idx_people_archived ON people(archived);
CREATE INDEX idx_organizations_archived ON organizations(archived);
CREATE INDEX idx_interactions_occurred_at ON interactions(occurred_at);
CREATE INDEX idx_interactions_type ON interactions(type) WHERE archived = 0;
CREATE INDEX idx_interaction_people_person_id ON interaction_people(person_id);
CREATE INDEX idx_deals_stage ON deals(stage) WHERE archived = 0;
CREATE INDEX idx_deals_person_id ON deals(person_id) WHERE archived = 0;
CREATE INDEX idx_deals_org_id ON deals(org_id) WHERE archived = 0;
CREATE INDEX idx_tasks_due_at ON tasks(due_at) WHERE completed = 0 AND archived = 0;
CREATE INDEX idx_tasks_person_id ON tasks(person_id) WHERE archived = 0;
CREATE INDEX idx_tasks_completed ON tasks(completed) WHERE archived = 0;
CREATE INDEX idx_taggings_entity ON taggings(entity_type, entity_id);
CREATE INDEX idx_taggings_tag_id ON taggings(tag_id);
CREATE INDEX idx_custom_fields_entity ON custom_fields(entity_type, entity_id);
CREATE INDEX idx_relationships_person_id ON relationships(person_id);
CREATE INDEX idx_relationships_related_person_id ON relationships(related_person_id);

-- down

DROP TABLE IF EXISTS relationships;
DROP TABLE IF EXISTS custom_fields;
DROP TABLE IF EXISTS taggings;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS deals;
DROP TABLE IF EXISTS interaction_people;
DROP TABLE IF EXISTS interactions;
DROP TABLE IF EXISTS people;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS people_fts;
DROP TABLE IF EXISTS organizations_fts;
DROP TABLE IF EXISTS interactions_fts;
DROP TABLE IF EXISTS deals_fts;
