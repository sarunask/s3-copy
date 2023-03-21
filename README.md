# s3-copy
Tool for faster transfer multiple files to AWS S3, by using GO routines.

## Install if you have go installed

```bash
go get github.com/sarunask/s3-copy
```

## Usage

You will need setup your AWS access first. Please do one of the bellow:
1. If you run on EC2 instance with IAM role, nothing is required to do
1. If you run on your own PC, please setup ~/.aws/ with `aws configure` command
1. You can export 2 env variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` which will grant you access to AWS

Dry run, with debug, nothing uploaded. You could check if exclude works as expected.
```bash
./s3-copy --path ~/Music/ --s3-bucket=ssss --dry-run --debug --sse-c-key 45123qwefawdfgddddadfqwefgqwegdd --exclude '.*\.mp4' --workers 50
```

If you want to use CSV file as input
```bash
./s3-copy --s3-bucket some-bucket --input-csv input.csv
```

CSV file format `localFileName,s3ObjectNameWithPath`:
```csv
../test/file1.bin,/customers/gu/upload/fileUp1.bin
../test/file2.bin,/customers/gu/upload/fileUp2.bin
```

CSV file supports a wildcard in the file name. In case of multiple entry the first (alphabetically) file will be used
```csv
../test/file1.*,/customers/gu/upload/file1.*
../test/file2.bin,/customers/gu/upload/fileUp2.bin
```
