![Build status](https://github.com/joonas-fi/tamperewebcam/workflows/Build/badge.svg)

Takes (and crops) image from one of the camera feeds in my hometown and makes it available
at a stable URL and archives the images for me to maybe make automated timelapse videos later.

Runs in AWS Lambda


Published image at
------------------

https://s3.amazonaws.com/files.function61.com/tampere-webcam/hiedanranta/latest.jpg


IAM policy
----------

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "putWebcamImages",
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:CopyObject",
                "s3:PutObjectAcl"
            ],
            "Resource": [
                "arn:aws:s3:::files.function61.com/tampere-webcam/*"
            ]
        }
    ]
}
```
