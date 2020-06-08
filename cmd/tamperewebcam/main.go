package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
	"github.com/function61/gokit/aws/lambdautils"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/gokit/osutil"
	"github.com/spf13/cobra"
)

func main() {
	if lambdautils.InLambda() {
		lambda.StartHandler(&simpleFnLambdaAdapter{run})
		return
	}

	app := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Tampere webcam",
		Version: dynversion.Version,
	}

	app.AddCommand(runEntry())

	osutil.ExitIfError(app.Execute())
}

func run(ctx context.Context) error {
	bucketCtx, err := Bucket("files.function61.com", "us-east-1")
	if err != nil {
		return err
	}

	// 14:50 picture came online 14:53, so better schedule this at:
	// - 14:07 to access 14:00
	// - 14:17 to access 14:10
	// - etc
	ts := floorTenMinutes(time.Now())

	croppedFile, err := obtainCroppedImage(ctx, ts)
	if err != nil {
		return err
	}

	longTermKey := fmt.Sprintf("tampere-webcam/hiedanranta/%s.jpg", ts.UTC().Format("2006-01-02 15-04-05Z"))
	latestKey := fmt.Sprintf("tampere-webcam/hiedanranta/%s", "latest.jpg")

	log.Printf("uploading to %s", longTermKey)

	if _, err := bucketCtx.S3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      bucketCtx.Name,
		Key:         aws.String(longTermKey),
		ContentType: aws.String("image/jpeg"),
		Body:        bytes.NewReader(croppedFile.Bytes()), // TODO: more optimal?
	}); err != nil {
		return err
	}

	log.Println("making a copy to latest")

	// copy source includes source bucket name in front, and also for somet reason needs to
	// be URL escaped, while key does not
	copySource := url.PathEscape(fmt.Sprintf("%s/%s", *bucketCtx.Name, longTermKey))

	// metadata from source is copied, like content-type
	if _, err := bucketCtx.S3.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     bucketCtx.Name,
		Key:        aws.String(latestKey),
		CopySource: aws.String(copySource),
	}); err != nil {
		return err
	}

	return nil
}

func obtainCroppedImage(ctx context.Context, ts time.Time) (*bytes.Buffer, error) {
	imgUrl, err := roundshotHiedanranta.Url(ts, roundshotImageSizeVariantFull)
	if err != nil {
		return nil, err
	}

	log.Printf("requesting %s", imgUrl)

	imgResp, err := ezhttp.Get(ctx, imgUrl)
	if err != nil {
		return nil, err
	}
	imgBytes := imgResp.Body
	defer imgBytes.Close()

	log.Println("decoding & downloading image")
	img, err := jpeg.Decode(imgBytes)
	if err != nil {
		return nil, err
	}

	log.Println("cropping")

	cropped := imaging.Crop(img, image.Rect(8668, 0, 8668+1792, 1012))

	// need buffer b/c S3 client needs io.ReadSeeker
	croppedFile := &bytes.Buffer{}

	log.Println("encoding")

	if err := jpeg.Encode(croppedFile, cropped, nil); err != nil {
		return nil, err
	}

	return croppedFile, nil
}

func runEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Fetch & store webcam snapshot",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(run(
				osutil.CancelOnInterruptOrTerminate(nil)))
		},
	}
}

// 13:00 => 13:00
// 13:05 => 13:00
// 13:09 => 13:09
// 13:10 => 13:10
// ...
func floorTenMinutes(ts time.Time) time.Time {
	// 14:50 picture came online 14:53, so we should always go five minutes back
	return time.Date(
		ts.Year(),
		ts.Month(),
		ts.Day(),
		ts.Hour(),
		int(float64(ts.Minute())/10.0)*10,
		0,
		0,
		ts.Location())
}

// adapts a simple function (discards payload, returns empty payload) as a Lambda handler
type simpleFnLambdaAdapter struct {
	fn func(context.Context) error
}

func (h *simpleFnLambdaAdapter) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	return []byte{}, run(ctx)
}
