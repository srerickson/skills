# Agent Skills Repository

A repository of agent skills, organized with one skill per directory.

Skills provide specialized domain knowledge, operational patterns, and reference workflows for autonomous coding and engineering agents.

---

## Available Skills

| Skill | Directory | Description |
| :--- | :--- | :--- |
| **digital-preservation-processes** | [`digital-preservation-processes/`](./digital-preservation-processes/) | Digital preservation workflows and pipeline design for Trustworthy Digital Archives (TDAs) based on the EOSC EDEN Core Preservation Processes (CPPs). |

---

## Digital Preservation Processes Skill

The `digital-preservation-processes` skill provides an operational blueprint for implementing digital preservation workflows according to the 30 Core Preservation Processes (CPP-001 through CPP-030) formulated by the EOSC EDEN project.

### Key Components:
- **Conceptual Model**: Deconstructs information packages (SIP, AIP, DIP) and objects (Intellectual Entity, Representation, File, Bitstream, Environment) alongside administrative and descriptive metadata.
- **Process Catalog & References**: Comprehensive operational specifications for all 30 CPPs in [`digital-preservation-processes/references/`](./digital-preservation-processes/references/).
- **Workflow Patterns**: Canonical ingest pipelines (`CPP-029`), fixity checking schedules (`CPP-003`), data corruption remediation (`CPP-004`), and characterisation/normalisation sequences.

---

## References and Citations

The digital preservation skill and its accompanying references are derived from and cite the **Core Preservation Processes (CPPs)** created by the **EOSC EDEN project** (European Open Science Cloud - European Digital Ecosystem for Networked Data):

- **Project Website**: [https://eden-fidelis.eu/](https://eden-fidelis.eu/)
- **Core Preservation Processes Website**: [https://eosc-eden.github.io/wp1-cpp-descriptions/](https://eosc-eden.github.io/wp1-cpp-descriptions/)
- **GitHub Repository**: [EOSC-EDEN/wp1-cpp-descriptions](https://github.com/EOSC-EDEN/wp1-cpp-descriptions)
- **Interactive Visualisation Tool**: [https://eosc-eden.github.io/wp1-cpp-visualization](https://eosc-eden.github.io/wp1-cpp-visualization)
- **Conceptual Model & Glossary**:
  - EOSC-EDEN Task 1.2, *M1.1 Report on Identification of Core Preservation Processes: Glossary*, 2025-08-31. Available at: [https://github.com/EOSC-EDEN/wp1-cpp-descriptions/blob/main/03_Glossary.pdf](https://github.com/EOSC-EDEN/wp1-cpp-descriptions/blob/main/03_Glossary.pdf).
- **Milestone Citation (Zenodo)**:
  - EOSC EDEN T1.2, Micky Lindlar, Bertrand Caron, et al. *M1.1 Report on Identification of Core Preservation Processes*. Zenodo, 2025. [https://doi.org/10.5281/ZENODO.16992451](https://doi.org/10.5281/ZENODO.16992451).
- **Harmonized Standards**:
  - **OAIS**: ISO 14721:2012 / CCSDS 650.0-M-3 (*Reference Model for an Open Archival Information System*).
  - **PREMIS**: *PREMIS Data Dictionary for Preservation Metadata*, version 3.0 (Library of Congress).
  - **Trustworthy Repositories**: ISO 16363 / CCSDS 652.0-M-1; CoreTrustSeal Requirements; DIN 31644.
