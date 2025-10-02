
## QuickSight Integration

This tool uploads CSV files to S3 and maintains a cumulative manifest file that QuickSight can use to import all historical data.

### How It Works

1. Each cron run queries the database and saves a timestamped CSV file
2. The CSV is uploaded to S3: `s3://{bucket}/{prefix}/{report_name}/{timestamp}.csv`
3. A manifest file is updated locally to include all S3 URIs (cumulative)
4. The manifest is uploaded to S3: `s3://{bucket}/{prefix}/manifests/{report_name}/manifest.json`
5. QuickSight reads the manifest and imports all referenced CSV files, combining the data

### Manifest File Format

The generated manifest follows AWS QuickSight specifications:

```json
{
  "fileLocations": [
    {
      "URIs": [
        "s3://your-bucket/encode-reports/circulation_report/2025-01-15.02.00.00.csv",
        "s3://your-bucket/encode-reports/circulation_report/2025-01-16.02.00.00.csv",
        "s3://your-bucket/encode-reports/circulation_report/2025-01-17.02.00.00.csv"
      ]
    }
  ],
  "globalUploadSettings": {
    "format": "CSV",
    "delimiter": ",",
    "textqualifier": "\"",
    "containsHeader": "true"
  }
}
```

### Setting Up QuickSight

#### Prerequisites

1. **S3 Bucket Access**: Ensure QuickSight has read permissions to your S3 bucket
   - In AWS Console → QuickSight → Manage QuickSight → Security & permissions
   - Under "QuickSight access to AWS services", enable S3 and select your bucket

2. **Consistent Schema**: CSV files must have consistent column names and data types across runs
   - Column order must remain the same
   - QuickSight takes field names from the first file

#### Creating the Dataset

1. **Navigate to QuickSight**
   - Go to the Amazon QuickSight start page
   - Choose **Datasets**

2. **Create New Dataset**
   - Click **New dataset**
   - In the "FROM NEW DATA SOURCES" section, choose the **Amazon S3** icon

3. **Configure Data Source**
   - **Data source name**: Enter a descriptive name (e.g., "Circulation Report")
   - **Upload a manifest file**: Choose **URL**
   - **Manifest URL**: Enter the S3 URL for your manifest:
```
https://{bucket}.s3-{region}.amazonaws.com/{prefix}/manifests/{report_name}/manifest.json
```
Example:
```
https://library-reports.s3-us-east-1.amazonaws.com/encode-reports/manifests/circulation_report/manifest.json
```

4. **Connect and Verify**
   - Click **Connect**
   - Choose **Edit/Preview data** to verify the data loaded correctly
   - QuickSight will combine all CSV files listed in the manifest

5. **Schedule Refresh** (Optional)
   - Go to Datasets → Select your dataset → Refresh → Schedule refresh
   - Set refresh schedule to match or exceed your cron schedule
   - Each refresh will automatically include new CSV files added to the manifest

### Important Notes

- **Schema Consistency**: All CSV files must have the same columns in the same order
- **Cumulative Data**: The manifest grows over time, including all historical CSV files
- **Automatic Updates**: When QuickSight refreshes, it automatically picks up new files from the manifest
- **No Manual Intervention**: Once configured, QuickSight automatically imports new data on each refresh

### Troubleshooting

**QuickSight can't access the manifest:**
- Verify QuickSight has S3 permissions for your bucket
- Check the manifest URL is publicly accessible or QuickSight has proper IAM role

**Data not updating:**
- Verify the cron schedule is running (`./encode run --log-level DEBUG`)
- Check S3 for new CSV files and updated manifest
- Ensure QuickSight refresh schedule is set correctly

**Column mismatch errors:**
- All CSV files must have identical column names and order
- Check your SQL queries maintain consistent schema across runs
