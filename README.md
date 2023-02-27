# s3-copy
Tool for faster transfer multiple files to AWS S3, by using GO routines.

## Install if you have go installed

```bash
go get github.com/sarunask/s3-copy
```

## Usage

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