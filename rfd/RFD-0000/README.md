---
authors: Joe Corall
state: draft
---

# RFD 0000 - Creating a data pipeline

## Required Approvers

* Library Technology Team

## What

Build a pipeline to send data from library systems to AWS Quicksight.

## Why

Lehigh Library & Technology Services is standardizing its data reporting for Lehigh administrators by using AWS Quicksight as a central data hub. This gives a one stop shop when questions arise that can be answered by data we collect.

## What data will be sent to Quicksight?

Data intended for wider sharing, including:

- Inter-departmental or unit reports
- Annual reports
- Public dashboards

If your current system meets your reporting needs, continue using its built-in capabilities. Only export data to Quicksight when people who can't access the source data require it or when enhanced reporting is needed.

## How

Library Technology can build automation to automatically perform the extraction, transform, and loading on a set schedule. If automation is not possible, it will require placing a CSV file somewhere on the dlshare periodically or some other manual method to get the CSV into Quicksight.

### Extract

- Export data from a given system (e.g. Google Analytics, LibAnswers, MetaDB, etc) in CSV format
  - If your data lives in Google Sheets - your job is done! Assumming your data is consistent and your columns will remain stable
  - Formats other than CSV can be supported, but CSV is the happy path.
- Generate one CSV per AWS Quicksight dataset.

### Transform

- CSVs will be uploaded to their respective Google Sheet to allow appending data on the harvest/ETL schedule.
  - AWS Quicksight works best with a single consolidated file rather than multiple incremental uploads.
  - **TODO:** Confirm this approach with Jim Monek. We have flexibility here and could use the help of prior art to inform our implementation.

#### Integrity check

- Before appending a CSV to its Google Sheet, verify the CSV columns:
  - Are in the same order
    - New columns can be added, but existing columns must persist
      - this could be reconciled outside of the automation, and we should build docs on how to do that. But reconciliation will be a manual process
  - Each column has the expected data type (and optionally they are within a range)
  - Allow for controlled vocabularies within a column
  - Other validation techniques as we see fit during implementation
- If checks fail, alert the dataset owner and halt the process to protect Quicksight visualizations from breaking.

### Load

- Convert the Google Sheet back into a consolidated CSV (this contains all historical data and the latest append).
- Upload the CSV to an S3 bucket.
- AWS Quicksight will use S3 as the data source, so any existing dashboards will get their data refreshed.

### Visualize

After the data is loaded into Quicksight, we can build dashboards within Quicksight to visualize the data.

### Disaster Recovery

Since we'll be uploading to s3 we'll have our data stored in both Google Sheets and S3 which gives us great redundancy should one service have some disaster. S3's versioning enables point-in-time restores, mitigating the risk of erroneous data ingestions corrupting our data and going undetected until someone identifies the problem. Although Google Sheets also offers versioning and point-in-time recovery, its functionality should be regarded as a backup or disaster recovery option to S3's more reliable system.

### Original discussion

The original GitHub discussion can be viewed at https://github.com/lehigh-university-libraries/scratchpad/pull/13
