---
name: digital-preservation-processes
description: Use when designing, evaluating, implementing, or auditing digital preservation workflows and trustworthy digital archive (TDA) pipelines based on the EOSC EDEN Core Preservation Processes (CPPs).
---

# Digital Preservation Processes Skill

A specialized skill for designing, structuring, evaluating, and implementing digital preservation workflows and pipelines for Trustworthy Digital Archives (TDAs), grounded in the **Core Preservation Processes (CPPs)** developed by the European Open Science Cloud (EOSC) EDEN project.

---

## 1. Scope & Purpose

This skill guides the implementation of operational digital preservation workflows across the entire digital archiving lifecycle:
- Ingest processing, pre-ingest validation, quarantine, virus scanning, and SIP packaging.
- Bit-level preservation, checksum generation/validation, fixity monitoring, storage media refreshment, and replication.
- Preservation characterisation: format identification, metadata extraction, format validation, and quality assessment.
- Preservation actions: file normalisation, migration, repair, emulation environment management, and derivative creation.
- Archival management: AIP versioning, identifier management (PIDs), batch export, exit scenarios, and disposal/tombstoning.
- Access & discovery: search index preparation, rights assessment, access delivery, and DIP generation.

Each Core Preservation Process (CPP-001 through CPP-030) is documented in full operational detail in the [references/](./references/) directory.

---

## 2. Conceptual Model (EOSC EDEN CPP)

The CPP conceptual model establishes a unified terminology and entity hierarchy harmonizing the **OAIS Reference Model (ISO 14721 / CCSDS 650.0-M-3)** and the **PREMIS Data Dictionary for Preservation Metadata (v3.0)**.

### 2.1 The Three Core High-Level Entities

```
+---------------------------------------------------------------------------------+
|                              Information Package                                |
|   (SIP: Submission, AIP: Archival [as preserved], DIP: Dissemination [access])  |
|                                                                                 |
|  +-------------------------------------+  +----------------------------------+  |
|  |        (Information) Object         |  |             Metadata             |  |
|  |                                     |  |                                  |  |
|  | +---------------------------------+ |  | +------------------------------+ |  |
|  | |      Intellectual Entity        | |  | |    Descriptive Metadata      | |  |
|  | | (Coherent unit described as one)| |  | | (Context, discovery, title,   | |  |
|  | +---------------------------------+ |  | |  creators, subjects)          | |  |
|  |                 |                   |  | +------------------------------+ |  |
|  |                 v                   |  | +------------------------------+ |  |
|  | +---------------------------------+ |  | |    Administrative Metadata   | |  |
|  | |        Representation           | |  | | (Management through processes) | |
|  | | (Files + Structural metadata)   | |  | | - Fixity Metadata (checksums)  | |
|  | +---------------------------------+ |  | | - Technical Metadata (codecs) | |  |
|  |                 |                   |  | | - Provenance Metadata (events)| |  |
|  |                 v                   |  | | - Rights Metadata (statements)| |  |
|  | +---------------------------------+ |  | | - Structural Metadata (links) | |  |
|  | |             File(s)             | |  | +------------------------------+ |  |
|  | | (Byte sequences in filesystem)  | |  +----------------------------------+  |
|  | +---------------------------------+ |                                        |
|  |                 |                   |                                        |
|  |                 v                   |                                        |
|  | +---------------------------------+ |                                        |
|  | |           Bitstream(s)          | |                                        |
|  | | (Contiguous/meaningful streams) | |                                        |
|  | +---------------------------------+ |                                        |
|  +-------------------------------------+                                        |
+---------------------------------------------------------------------------------+
```

#### 1. (Information) Object
The entity that must be preserved as a priority. Digital in nature, decomposed per PREMIS levels:
- **Intellectual Entity**: A coherent set of content described as a unit (e.g., a book, dataset, video, photograph, website).
- **Representation**: A set of stored files and structural metadata needed to provide a complete rendition of an Intellectual Entity (e.g., TIFF master representation vs. PDF access representation).
- **File**: A named, ordered sequence of bytes known by an operating system with file system statistics and a file format.
- **Bitstream**: Contiguous or non-contiguous data within a file with meaningful common properties requiring independent description or processing (e.g., audio/video streams in an MP4, embedded ICC profiles, thumbnail bitstreams).
- **Environment Object**: Software, operating system, and hardware stack specifications preserved to render an Object via emulation or virtualisation.

#### 2. Metadata
Structured, machine-actionable textual information describing preserved Objects:
- **Descriptive Metadata**: Enables discovery, identification, and understanding of context (Dublin Core, DataCite, MODS, etc.).
- **Administrative Metadata**: Operates preservation processes, subdivided into:
  - *Technical Metadata*: Characteristics of files/bitstreams (codecs, color spaces, sample rates, format validity).
  - *Fixity Metadata*: Cryptographic hashes and algorithms capturing bitwise states (`CPP-001`).
  - *Provenance Metadata*: Audit logs and event histories documenting all actions, agents, software, and timestamps (`CPP-016`).
  - *Rights Metadata*: Machine-actionable rights statements, copyright status, licences, and access constraints (`CPP-020`).
  - *Structural Metadata*: Maps internal relationships between files, representations, and intellectual entities.

#### 3. Information Package (IP)
Conceptual container gathering Objects and associated Metadata:
- **Submission Information Package (SIP)**: Delivered by producer or submitting system to the TDA (`CPP-029`).
- **Archival Information Package (AIP)**: Formatted, encapsulated, and versioned package preserved in TDA storage (`CPP-006`, `CPP-021`).
- **Dissemination Information Package (DIP)**: Derived from AIPs and provided to consumers/discovery interfaces (`CPP-025`, `CPP-028`).

---

## 3. Trustworthy Digital Archive (TDA) & Policy Hierarchy

A **Trustworthy Digital Archive (TDA)** consists of hardware, software, organization, and policy-governed procedures operating CPPs to fulfill its preservation mission.

Workflows must be parameterized by specific TDA policies:
- **Packaging Policy**: Defines directory schemas, manifest formats (e.g., BagIt, METS, OCFL), and checksum file paths (`CPP-021`, `CPP-029`).
- **Storage Management Policies**:
  - *Checksum Algorithms*: Accepted and required hashing algorithms (e.g., SHA-256, SHA-512) (`CPP-001`, `CPP-003`).
  - *Copies*: Quantity and diversity of active AIP replicas (`CPP-011`).
  - *Integrity Checking*: Frequency and schedule of fixity audits (`CPP-003`).
  - *Media*: Management of storage media lifecycles and media migrations (`CPP-030`).
- **File Format Policies**:
  - *Identification*: Selected identification tools and registries (PRONOM / DROID, Siegfried, Apache Tika) (`CPP-008`).
  - *Preferred Formats*: Target preservation and distribution formats (`CPP-018`, `CPP-026`).
  - *Validation*: Validators and error tolerances (JHOVE, VeraPDF, Mediaconch) (`CPP-010`).
  - *Derivatives*: Transformation recipes for access copies (`CPP-028`).
- **Preservation Action Plans & Registry**: Concrete pathways, tools, and parameters for migration and repair (`CPP-012`, `CPP-014`, `CPP-027`).
- **Access & Rights Policies**: Authenticated access levels, embargoes, and licensing rules (`CPP-020`, `CPP-025`).
- **Retention & Exit Scenario Plans**: Retention periods, deaccessioning/disposal (`CPP-017`), and wholesale repository migration plans (`CPP-006`).

---

## 4. CPP Process Relationships

Processes in digital preservation workflows interact via formal relationship types:

| Relationship | Inverse | Definition & Workflow Impact |
| :--- | :--- | :--- |
| **Requires** | *Required by* | **Hard dependency**: Target CPP must have completed to allow execution (e.g., `CPP-003 Integrity Checking` requires `CPP-001 Checksum Generation`). |
| **May require** | *May be required by* | Conditional dependency depending on policy or object type. |
| **Triggers** | *Triggered by* | **Procedural invocation**: An event or error in one process initiates another (e.g., format error in `CPP-010` triggers `CPP-027 File Repair` or quarantine). |
| **Supplies** | *Customer* | Output of one step provides direct input to a subsequent process step. |
| **Affects** | *Affected by* | Changes behavior or data structures of another process (e.g., `CPP-021 AIP Versioning` affects `CPP-006 AIP Batch Export`). |
| **Facilitates** | *Facilitated by* | Makes execution easier or more reliable without being strictly mandatory. |
| **Alternative to** | *Alternative to* | Strategy choice (e.g., `CPP-014 File Migration` vs. `CPP-015 Emulation and Rendering Tools`). |
| **Affinity with** | *Affinity with* | Conceptual similarity (e.g., `CPP-014 File Migration` and `CPP-027 File Repair`). |
| **Not to be confused with** | *Not to be confused with* | Clear boundary distinction (e.g., Migration transforms the primary preservation representation; Derivatives produce secondary dissemination copies). |

---

## 5. Catalog of Core Preservation Processes (CPP-001 – CPP-030)

Detailed operational specifications for each process are located in `references/`:

| Identifier | Process Label | Core Function | Reference File |
| :--- | :--- | :--- | :--- |
| **CPP-001** | Checksum Generation and Recording | Computes and registers cryptographic message digests for files | [CPP-001.md](./references/CPP-001.md) |
| **CPP-002** | Checksum Validation | Recalculates and verifies hashes against recorded fixity metadata | [CPP-002.md](./references/CPP-002.md) |
| **CPP-003** | Integrity Checking | Automated scheduled audits of repository storage against fixity logs | [CPP-003.md](./references/CPP-003.md) |
| **CPP-004** | Data Corruption Management | Escalation, quarantine, root-cause analysis, and restoration from replica | [CPP-004.md](./references/CPP-004.md) |
| **CPP-005** | Identifier Management | Mints, updates, resolves, and manages persistent identifiers (PIDs) | [CPP-005.md](./references/CPP-005.md) |
| **CPP-006** | AIP Batch Export | Extracts and packages AIPs for system exit, disaster recovery, or exchange | [CPP-006.md](./references/CPP-006.md) |
| **CPP-007** | Virus Scanning | Scans incoming files for malware and handles quarantine procedures | [CPP-007.md](./references/CPP-007.md) |
| **CPP-008** | File Format Identification | Matches file signatures/magic numbers against registries (PRONOM/PUID) | [CPP-008.md](./references/CPP-008.md) |
| **CPP-009** | Metadata Extraction | Extracts embedded technical and structural metadata from files | [CPP-009.md](./references/CPP-009.md) |
| **CPP-010** | File Format Validation | Validates syntactical conformance against formal format specifications | [CPP-010.md](./references/CPP-010.md) |
| **CPP-011** | Replication | Distributes identical AIP replicas across distinct storage systems/media | [CPP-011.md](./references/CPP-011.md) |
| **CPP-012** | Risk Mitigation | Formulates preservation actions to counteract obsolescence or decay | [CPP-012.md](./references/CPP-012.md) |
| **CPP-013** | Object Management Reporting | Aggregates repository health, format risk profiles, and capacity reports | [CPP-013.md](./references/CPP-013.md) |
| **CPP-014** | File Migration | Converts files from obsolete/at-risk formats to sustainable targets | [CPP-014.md](./references/CPP-014.md) |
| **CPP-015** | Emulation and Rendering Tools | Maintains software environments and emulators for authentic playback | [CPP-015.md](./references/CPP-015.md) |
| **CPP-016** | Metadata Ingest and Management | Manages schema mapping, provenance events, and metadata persistence | [CPP-016.md](./references/CPP-016.md) |
| **CPP-017** | Disposal | Permanently purges deaccessioned AIPs and records tombstone metadata | [CPP-017.md](./references/CPP-017.md) |
| **CPP-018** | Community Watch | Monitors external technology, format obsolescence, and community standards | [CPP-018.md](./references/CPP-018.md) |
| **CPP-019** | Data Quality Assessment | Evaluates incoming submissions against archival policies and criteria | [CPP-019.md](./references/CPP-019.md) |
| **CPP-020** | Rights Management | Validates licences, copyright permissions, and access constraints | [CPP-020.md](./references/CPP-020.md) |
| **CPP-021** | AIP Versioning | Tracks changes, creates incremental/new AIP editions, updates manifests | [CPP-021.md](./references/CPP-021.md) |
| **CPP-022** | Significant Properties Definition | Identifies properties (visual, layout, behavior) that must be preserved | [CPP-022.md](./references/CPP-022.md) |
| **CPP-023** | Risk Properties Definition & Extraction| Detects high-risk features (fonts, macros, encryption, proprietary codecs)| [CPP-023.md](./references/CPP-023.md) |
| **CPP-024** | Enabling Discovery | Indexes metadata, registers OAI-PMH endpoints, and exposes catalogs | [CPP-024.md](./references/CPP-024.md) |
| **CPP-025** | Enabling Access | Authenticates users, validates rights, and delivers DIP packages | [CPP-025.md](./references/CPP-025.md) |
| **CPP-026** | File Normalisation | Converts incoming non-preferred formats into preservation masters | [CPP-026.md](./references/CPP-026.md) |
| **CPP-027** | File Repair | Repaired corrupt file headers, bitstreams, or packaging defects | [CPP-027.md](./references/CPP-027.md) |
| **CPP-028** | Creation of Derivatives | Generates web-friendly, downsampled access copies (e.g., JPEG, MP4, PDF) | [CPP-028.md](./references/CPP-028.md) |
| **CPP-029** | Ingest | Orchestrates the end-to-end transformation of an SIP into an AIP | [CPP-029.md](./references/CPP-029.md) |
| **CPP-030** | Refreshment | Migrates data onto new storage media to prevent physical media failure | [CPP-030.md](./references/CPP-030.md) |

---

## 6. Digital Preservation Workflow Implementation Patterns

When designing or implementing preservation pipelines, follow these standard orchestration patterns:

### Pattern A: Standard Ingest Pipeline (`CPP-029`)
```
SIP Submission
      │
      ▼
1. Virus Scanning (CPP-007) ──[Infected]──► Quarantine & Notification
      │ [Clean]
      ▼
2. Checksum Validation (CPP-002) (Verify producer checksums)
      │
      ▼
3. Characterisation:
   ├─ Format Identification (CPP-008) (e.g. Siegfried/DROID -> PUID)
   ├─ Metadata Extraction (CPP-009) (e.g. ExifTool, Apache Tika, FITS)
   └─ Format Validation (CPP-010) (e.g. JHOVE, VeraPDF)
      │
      ▼
4. Data Quality & Rights Assessment (CPP-019, CPP-020)
      │
      ▼
5. Optional Preservation Actions:
   ├─ File Normalisation (CPP-026) (to preferred format)
   └─ Creation of Derivatives (CPP-028) (for access)
      │
      ▼
6. Package & Register:
   ├─ Checksum Generation (CPP-001) (New hashes for transformed/preservation files)
   ├─ Identifier Management (CPP-005) (Mint PID / Archival ID)
   ├─ Metadata & Provenance Logging (CPP-016) (PREMIS events)
   └─ AIP Assembly & Storage (CPP-021)
      │
      ▼
7. Replication (CPP-011) (Disperse across redundant storage nodes)
      │
      ▼
8. Enable Discovery & Access (CPP-024, CPP-025)
```

### Pattern B: Scheduled Routine Maintenance Pipeline
```
Cron / Scheduler
      │
      ▼
1. Storage Integrity Checking (CPP-003)
      │
      ├──[Mismatch/Corruption]──► Data Corruption Management (CPP-004)
      │                                └── Restore from replica (CPP-011)
      ▼
2. Media Refreshment Check (CPP-030)
      │
      └──[Media Aging]──────────► Migrate bytes to new media
      ▼
3. Format Risk & Obsolescence Audit (CPP-018, CPP-023, CPP-013)
      │
      └──[Format at Risk]───────► Risk Mitigation Planning (CPP-012)
                                       └── File Migration (CPP-014)
```

---

## 7. How to Use this Skill

When tasked with implementing, auditing, or designing digital preservation workflows:
1. **Identify the Core Process**: Locate the matching CPP number(s) from the catalog.
2. **Consult the Reference**: Read the corresponding file in `references/CPP-XXX.md` for explicit input/output schemas, required policy configurations, trigger events, step sequences, and customer/supplier dependencies.
3. **Model the Data Entities**: Adhere to the PREMIS/OAIS conceptual model (SIP -> AIP -> DIP, Intellectual Entity -> Representation -> File -> Bitstream).
4. **Document Provenance**: Record PREMIS-compliant events for all mutations (checksum calculation, normalisation, format validation, replication).
5. **Enforce Trustworthy Archival Standards**: Verify alignment with standards cited in the references (ISO 16363, CoreTrustSeal, OAIS ISO 14721).

---

## 8. Citations & Attribution

This skill and its references are based on the **Core Preservation Processes (CPPs)** created by the **EOSC EDEN project** (Horizon Europe Project Number 101188015).

### Primary References:
1. **EOSC EDEN CPP Descriptions**:
   - Website: https://eosc-eden.github.io/wp1-cpp-descriptions/
   - GitHub Repository: https://github.com/EOSC-EDEN/wp1-cpp-descriptions
   - Visualisation Tool: https://eosc-eden.github.io/wp1-cpp-visualization
2. **Glossary and Conceptual Model Report**:
   - EOSC-EDEN Task 1.2, *M1.1 Report on Identification of Core Preservation Processes: Glossary*, August 31, 2025.
   - Available at: https://github.com/EOSC-EDEN/wp1-cpp-descriptions/blob/main/03_Glossary.pdf
3. **Zenodo Milestone Publication**:
   - EOSC EDEN T1.2, Micky Lindlar, Bertrand Caron, et al., *M1.1 Report on Identification of Core Preservation Processes*, Zenodo, 2025. https://doi.org/10.5281/ZENODO.16992451
4. **Underlying Standards**:
   - OAIS: ISO 14721:2012 / CCSDS 650.0-M-3 (Reference Model for an Open Archival Information System).
   - PREMIS: Data Dictionary for Preservation Metadata, Version 3.0 (Library of Congress).
   - Trustworthy Repositories: CoreTrustSeal Requirements; ISO 16363 / CCSDS 652.0-M-1; DIN 31644 (Nestor Seal).
