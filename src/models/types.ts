import { z } from 'zod'

// --- People ---

export const PersonRow = z.object({
  id: z.number(),
  uuid: z.string(),
  first_name: z.string(),
  last_name: z.string().nullable(),
  email: z.string().nullable(),
  phone: z.string().nullable(),
  title: z.string().nullable(),
  company: z.string().nullable(),
  location: z.string().nullable(),
  linkedin: z.string().nullable(),
  twitter: z.string().nullable(),
  website: z.string().nullable(),
  notes: z.string().nullable(),
  summary: z.string().nullable(),
  org_id: z.number().nullable(),
  created_at: z.string(),
  updated_at: z.string(),
  archived: z.number(),
})

export type PersonRow = z.infer<typeof PersonRow>

export const PersonInsert = z.object({
  first_name: z.string().min(1),
  last_name: z.string().nullable().optional(),
  email: z.string().nullable().optional(),
  phone: z.string().nullable().optional(),
  title: z.string().nullable().optional(),
  company: z.string().nullable().optional(),
  location: z.string().nullable().optional(),
  linkedin: z.string().nullable().optional(),
  twitter: z.string().nullable().optional(),
  website: z.string().nullable().optional(),
  notes: z.string().nullable().optional(),
  summary: z.string().nullable().optional(),
  org_id: z.number().nullable().optional(),
})

export type PersonInsert = z.infer<typeof PersonInsert>

export const PersonUpdate = PersonInsert.partial()

export type PersonUpdate = z.infer<typeof PersonUpdate>

// --- Organizations ---

export const OrganizationRow = z.object({
  id: z.number(),
  uuid: z.string(),
  name: z.string(),
  domain: z.string().nullable(),
  industry: z.string().nullable(),
  location: z.string().nullable(),
  notes: z.string().nullable(),
  summary: z.string().nullable(),
  created_at: z.string(),
  updated_at: z.string(),
  archived: z.number(),
})

export type OrganizationRow = z.infer<typeof OrganizationRow>

export const OrganizationInsert = z.object({
  name: z.string().min(1),
  domain: z.string().nullable().optional(),
  industry: z.string().nullable().optional(),
  location: z.string().nullable().optional(),
  notes: z.string().nullable().optional(),
  summary: z.string().nullable().optional(),
})

export type OrganizationInsert = z.infer<typeof OrganizationInsert>

export const OrganizationUpdate = OrganizationInsert.partial()

export type OrganizationUpdate = z.infer<typeof OrganizationUpdate>

// --- Interactions ---

export const INTERACTION_TYPES = ['call', 'email', 'meeting', 'note', 'message'] as const
export type InteractionType = (typeof INTERACTION_TYPES)[number]

export const DIRECTIONS = ['inbound', 'outbound'] as const
export type Direction = (typeof DIRECTIONS)[number]

export const InteractionRow = z.object({
  id: z.number(),
  uuid: z.string(),
  type: z.enum(INTERACTION_TYPES),
  subject: z.string().nullable(),
  content: z.string().nullable(),
  direction: z.enum(DIRECTIONS).nullable(),
  occurred_at: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
  archived: z.number(),
})

export type InteractionRow = z.infer<typeof InteractionRow>

export const InteractionInsert = z.object({
  type: z.enum(INTERACTION_TYPES),
  subject: z.string().nullable().optional(),
  content: z.string().nullable().optional(),
  direction: z.enum(DIRECTIONS).nullable().optional(),
  occurred_at: z.string().optional(),
})

export type InteractionInsert = z.infer<typeof InteractionInsert>

// --- Deals ---

export const DEAL_STAGES = ['lead', 'prospect', 'proposal', 'negotiation', 'won', 'lost'] as const
export type DealStage = (typeof DEAL_STAGES)[number]

export const DealRow = z.object({
  id: z.number(),
  uuid: z.string(),
  title: z.string(),
  value: z.number().nullable(),
  currency: z.string().nullable(),
  stage: z.enum(DEAL_STAGES),
  person_id: z.number().nullable(),
  org_id: z.number().nullable(),
  closed_at: z.string().nullable(),
  notes: z.string().nullable(),
  created_at: z.string(),
  updated_at: z.string(),
  archived: z.number(),
})

export type DealRow = z.infer<typeof DealRow>

export const DealInsert = z.object({
  title: z.string().min(1),
  value: z.number().nullable().optional(),
  currency: z.string().optional(),
  stage: z.enum(DEAL_STAGES).optional(),
  person_id: z.number().nullable().optional(),
  org_id: z.number().nullable().optional(),
  closed_at: z.string().nullable().optional(),
  notes: z.string().nullable().optional(),
})

export type DealInsert = z.infer<typeof DealInsert>

export const DealUpdate = DealInsert.partial()

export type DealUpdate = z.infer<typeof DealUpdate>

// --- Tasks ---

export const PRIORITIES = ['low', 'normal', 'high', 'urgent'] as const
export type Priority = (typeof PRIORITIES)[number]

export const TaskRow = z.object({
  id: z.number(),
  uuid: z.string(),
  title: z.string(),
  description: z.string().nullable(),
  due_at: z.string().nullable(),
  priority: z.enum(PRIORITIES),
  completed: z.number(),
  completed_at: z.string().nullable(),
  person_id: z.number().nullable(),
  deal_id: z.number().nullable(),
  created_at: z.string(),
  updated_at: z.string(),
  archived: z.number(),
})

export type TaskRow = z.infer<typeof TaskRow>

export const TaskInsert = z.object({
  title: z.string().min(1),
  description: z.string().nullable().optional(),
  due_at: z.string().nullable().optional(),
  priority: z.enum(PRIORITIES).optional(),
  person_id: z.number().nullable().optional(),
  deal_id: z.number().nullable().optional(),
})

export type TaskInsert = z.infer<typeof TaskInsert>

export const TaskUpdate = TaskInsert.partial()

export type TaskUpdate = z.infer<typeof TaskUpdate>

// --- Tags ---

export const TagRow = z.object({
  id: z.number(),
  name: z.string(),
})

export type TagRow = z.infer<typeof TagRow>

export const TAGGABLE_ENTITIES = ['person', 'organization', 'deal', 'interaction'] as const
export type TaggableEntity = (typeof TAGGABLE_ENTITIES)[number]

export const TaggingRow = z.object({
  id: z.number(),
  tag_id: z.number(),
  entity_type: z.enum(TAGGABLE_ENTITIES),
  entity_id: z.number(),
})

export type TaggingRow = z.infer<typeof TaggingRow>

// --- Custom Fields ---

export const CustomFieldRow = z.object({
  id: z.number(),
  entity_type: z.string(),
  entity_id: z.number(),
  field_name: z.string(),
  field_value: z.string().nullable(),
})

export type CustomFieldRow = z.infer<typeof CustomFieldRow>

// --- Relationships ---

export const RELATIONSHIP_TYPES = [
  'colleague',
  'friend',
  'manager',
  'report',
  'mentor',
  'mentee',
  'referred-by',
  'referred',
] as const
export type RelationshipType = (typeof RELATIONSHIP_TYPES)[number]

export const RelationshipRow = z.object({
  id: z.number(),
  person_id: z.number(),
  related_person_id: z.number(),
  type: z.enum(RELATIONSHIP_TYPES),
  notes: z.string().nullable(),
  created_at: z.string(),
})

export type RelationshipRow = z.infer<typeof RelationshipRow>

export const RelationshipInsert = z.object({
  person_id: z.number(),
  related_person_id: z.number(),
  type: z.enum(RELATIONSHIP_TYPES),
  notes: z.string().nullable().optional(),
})

export type RelationshipInsert = z.infer<typeof RelationshipInsert>
