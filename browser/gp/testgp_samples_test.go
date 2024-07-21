package gp

import (
	"bytes"
	"encoding/json"
	"errors"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/psanford/memfs"
	"github.com/simulot/immich-go/immich/metadata"
)

type inMemFS struct {
	*memfs.FS
	err error
}

func newInMemFS() *inMemFS {
	return &inMemFS{
		FS: memfs.New(),
	}
}

func (mfs *inMemFS) addFile(name string, content []byte) *inMemFS {
	if mfs.err != nil {
		return mfs
	}
	dir := path.Dir(name)
	mfs.err = errors.Join(mfs.err, mfs.MkdirAll(dir, 0o777))
	mfs.err = errors.Join(mfs.err, mfs.WriteFile(name, content, 0o777))
	return mfs
}

func (mfs *inMemFS) addImage(name string, length int) *inMemFS {
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = byte(i % 256)
	}
	mfs.addFile(name, b)
	return mfs
}

type jsonFn func(md *GoogleMetaData)

func takenTime(date string) func(md *GoogleMetaData) {
	return func(md *GoogleMetaData) {
		md.PhotoTakenTime.Timestamp = strconv.FormatInt(metadata.TakeTimeFromName(date).Unix(), 10)
	}
}

func (mfs *inMemFS) addJSONImage(name string, title string, modifiers ...jsonFn) *inMemFS {
	md := GoogleMetaData{
		Metablock: Metablock{
			Title:      title,
			URLPresent: true,
		},
	}
	md.PhotoTakenTime.Timestamp = strconv.FormatInt(time.Date(2023, 10, 23, 15, 0, 0, 0, time.Local).Unix(), 10)
	for _, f := range modifiers {
		f(&md)
	}
	content := bytes.NewBuffer(nil)
	enc := json.NewEncoder(content)
	err := enc.Encode(md)
	if err != nil {
		panic(err)
	}
	mfs.addFile(name, content.Bytes())
	return mfs
}

func (mfs *inMemFS) addJSONAlbum(file string, albumName string) *inMemFS {
	return mfs.addFile(file, []byte(`{
		"title": "`+albumName+`",
		"description": "",
		"access": "",
		"date": {
		  "timestamp": "0",
		  "formatted": "1 janv. 1970, 00:00:00 UTC"
		},
		"location": "",
		"geoData": {
		  "latitude": 0.0,
		  "longitude": 0.0,
		  "altitude": 0.0,
		  "latitudeSpan": 0.0,
		  "longitudeSpan": 0.0
		}
	  }`))
}

type fileResult struct {
	name  string
	size  int
	title string
}

func sortFileResult(s []fileResult) []fileResult {
	sort.Slice(s, func(i, j int) bool {
		c := strings.Compare(s[i].name, s[j].name)
		switch {
		case c < 0:
			return true
		case c > 0:
			return false
		}
		return s[i].size < s[j].size
	})
	return s
}

func simpleYear() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2023/PXL_20230922_144936660.jpg.json", "PXL_20230922_144936660.jpg").
		addImage("Photos from 2023/PXL_20230922_144936660.jpg", 10).
		addJSONImage("Photos from 2023/PXL_20230922_144956000.jpg.json", "PXL_20230922_144956000.jpg").
		addImage("Photos from 2023/PXL_20230922_144956000.jpg", 20)
}

func simpleAlbum() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2020/IMG_8172.jpg.json", "IMG_8172.jpg", takenTime("20200101103000")).
		addImage("Photos from 2020/IMG_8172.jpg", 25).
		addJSONImage("Photos from 2023/PXL_20230922_144936660.jpg.json", "PXL_20230922_144936660.jpg", takenTime("PXL_20230922_144936660")).
		addJSONImage("Photos from 2023/PXL_20230922_144934440.jpg.json", "PXL_20230922_144934440.jpg", takenTime("PXL_20230922_144934440")).
		addJSONImage("Photos from 2023/IMG_8172.jpg.json", "IMG_8172.jpg", takenTime("20230922102100")).
		addImage("Photos from 2023/PXL_20230922_144936660.jpg", 10).
		addImage("Photos from 2023/PXL_20230922_144934440.jpg", 15).
		addImage("Photos from 2023/IMG_8172.jpg", 52).
		addJSONAlbum("Album/anyname.json", "Album").
		addJSONImage("Album/IMG_8172.jpg.json", "IMG_8172.jpg", takenTime("20230922102100")).
		addJSONImage("Album/PXL_20230922_144936660.jpg.json", "PXL_20230922_144936660.jpg", takenTime("PXL_20230922_144936660")).
		addImage("Album/IMG_8172.jpg", 52).
		addImage("Album/PXL_20230922_144936660.jpg", 10)
}

func albumWithoutImage() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2023/PXL_20230922_144936660.jpg.json", "PXL_20230922_144936660.jpg").
		addImage("Photos from 2023/PXL_20230922_144936660.jpg", 10).
		addJSONImage("Photos from 2023/PXL_20230922_144934440.jpg.json", "PXL_20230922_144934440.jpg").
		addImage("Photos from 2023/PXL_20230922_144934440.jpg", 15).
		addJSONAlbum("Album/anyname.json", "Album").
		addJSONImage("Album/PXL_20230922_144936660.jpg.json", "PXL_20230922_144936660.jpg").
		addImage("Album/PXL_20230922_144936660.jpg", 10).
		addJSONImage("Album/PXL_20230922_144934440.jpg.json", "PXL_20230922_144934440.jpg")
}

func namesWithNumbers() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2009/IMG_3479.JPG.json", "IMG_3479.JPG").
		addImage("Photos from 2009/IMG_3479.JPG", 10).
		addJSONImage("Photos from 2009/IMG_3479.JPG(1).json", "IMG_3479.JPG").
		addImage("Photos from 2009/IMG_3479(1).JPG", 12).
		addJSONImage("Photos from 2009/IMG_3479.JPG(2).json", "IMG_3479.JPG").
		addImage("Photos from 2009/IMG_3479(2).JPG", 15)
}

func namesTruncated() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2023/😀😃😄😁😆😅😂🤣🥲☺️😊😇🙂🙃😉😌😍🥰😘😗😙😚😋.json", "😀😃😄😁😆😅😂🤣🥲☺️😊😇🙂🙃😉😌😍🥰😘😗😙😚😋😛😝😜🤪🤨🧐🤓😎🥸🤩🥳😏😒😞😔😟😕🙁☹️😣😖😫😩🥺😢😭😤😠😡🤬🤯😳🥵🥶.jpg").
		addImage("Photos from 2023/😀😃😄😁😆😅😂🤣🥲☺️😊😇🙂🙃😉😌😍🥰😘😗😙😚😋😛.jpg", 10).
		addJSONImage("Photos from 2023/PXL_20230809_203449253.LONG_EXPOSURE-02.ORIGIN.json", "PXL_20230809_203449253.LONG_EXPOSURE-02.ORIGINAL.jpg").
		addImage("Photos from 2023/PXL_20230809_203449253.LONG_EXPOSURE-02.ORIGINA.jpg", 40).
		addJSONImage("Photos from 2023/05yqt21kruxwwlhhgrwrdyb6chhwszi9bqmzu16w0 2.jp.json", "05yqt21kruxwwlhhgrwrdyb6chhwszi9bqmzu16w0 2.jpg").
		addImage("Photos from 2023/05yqt21kruxwwlhhgrwrdyb6chhwszi9bqmzu16w0 2.jpg", 25)
}

func imagesEditedJSON() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2023/PXL_20220405_090123740.PORTRAIT.jpg.json", "PXL_20220405_090123740.PORTRAIT.jpg").
		addImage("Photos from 2023/PXL_20220405_090123740.PORTRAIT.jpg", 41).
		addImage("Photos from 2023/PXL_20220405_090123740.PORTRAIT-modifié.jpg", 21).
		addImage("Photos from 2023/PXL_20220405_090200110.PORTRAIT-modifié.jpg", 12)
}

func titlesWithForbiddenChars() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2012/27_06_12 - 1.mov.json", "27/06/12 - 1", takenTime("20120627")).
		addImage("Photos from 2012/27_06_12 - 1.mov", 52).
		addJSONImage("Photos from 2012/27_06_12 - 2.json", "27/06/12 - 2", takenTime("20120627")).
		addImage("Photos from 2012/27_06_12 - 2.jpg", 24)
}

func namesIssue39() *inMemFS {
	return newInMemFS().
		addJSONAlbum("Album/anyname.json", "Album").
		addJSONImage("Album/Backyard_ceremony_wedding_photography_xxxxxxx_.json", "Backyard_ceremony_wedding_photography_xxxxxxx_magnoliastudios-371.jpg", takenTime("20200101")).
		addJSONImage("Album/Backyard_ceremony_wedding_photography_xxxxxxx_(1).json", "Backyard_ceremony_wedding_photography_xxxxxxx_magnoliastudios-181.jpg", takenTime("20200101")).
		addJSONImage("Album/Backyard_ceremony_wedding_photography_xxxxxxx_(494).json", "Backyard_ceremony_wedding_photography_markham_magnoliastudios-19.jpg", takenTime("20200101")).
		addImage("Album/Backyard_ceremony_wedding_photography_xxxxxxx_m.jpg", 1).
		addImage("Album/Backyard_ceremony_wedding_photography_xxxxxxx_m(1).jpg", 181).
		addImage("Album/Backyard_ceremony_wedding_photography_xxxxxxx_m(494).jpg", 494).
		addJSONImage("Photos from 2020/Backyard_ceremony_wedding_photography_xxxxxxx_.json", "Backyard_ceremony_wedding_photography_xxxxxxx_magnoliastudios-371.jpg", takenTime("20200101")).
		addJSONImage("Photos from 2020/Backyard_ceremony_wedding_photography_xxxxxxx_(1).json", "Backyard_ceremony_wedding_photography_xxxxxxx_magnoliastudios-181.jpg", takenTime("20200101")).
		addJSONImage("Photos from 2020/Backyard_ceremony_wedding_photography_xxxxxxx_(494).json", "Backyard_ceremony_wedding_photography_markham_magnoliastudios-19.jpg", takenTime("20200101")).
		addImage("Photos from 2020/Backyard_ceremony_wedding_photography_xxxxxxx_m(1).jpg", 181).
		addImage("Photos from 2020/Backyard_ceremony_wedding_photography_xxxxxxx_m(494).jpg", 494).
		addImage("Photos from 2020/Backyard_ceremony_wedding_photography_xxxxxxx_m.jpg", 1)
}

func issue68MPFiles() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2022/PXL_20221228_185930354.MP.jpg.json", "PXL_20221228_185930354.MP.jpg", takenTime("20220101")).
		addImage("Photos from 2022/PXL_20221228_185930354.MP", 1).
		addImage("Photos from 2022/PXL_20221228_185930354.MP.jpg", 2)
}

func issue68LongExposure() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2023/PXL_20230814_201154491.LONG_EXPOSURE-01.COVER..json", "PXL_20230814_201154491.LONG_EXPOSURE-01.COVER.jpg", takenTime("20230101")).
		addImage("Photos from 2023/PXL_20230814_201154491.LONG_EXPOSURE-01.COVER.jpg", 1).
		addJSONImage("Photos from 2023/PXL_20230814_201154491.LONG_EXPOSURE-02.ORIGIN.json", "PXL_20230814_201154491.LONG_EXPOSURE-02.ORIGINAL.jpg", takenTime("20230101")).
		addImage("Photos from 2023/PXL_20230814_201154491.LONG_EXPOSURE-02.ORIGINA.jpg", 2)
}

func issue68ForgottenDuplicates() *inMemFS {
	return newInMemFS().
		addJSONImage("Photos from 2022/original_1d4caa6f-16c6-4c3d-901b-9387de10e528_.json", "original_1d4caa6f-16c6-4c3d-901b-9387de10e528_PXL_20220516_164814158.jpg", takenTime("20220101")).
		addImage("Photos from 2022/original_1d4caa6f-16c6-4c3d-901b-9387de10e528_P.jpg", 1).
		addImage("Photos from 2022/original_1d4caa6f-16c6-4c3d-901b-9387de10e528_P(1).jpg", 2)
}

// #390 Question: report shows way less images uploaded than scanned
func issue390WrongCount() *inMemFS {
	return newInMemFS().
		addJSONImage("Takeout/Google Photos/Photos from 2021/image000000.jpg.json", "image000000.jpg").
		addJSONImage("Takeout/Google Photos/Photos from 2021/image000000.jpg(1).json", "image000000.jpg").
		addJSONImage("Takeout/Google Photos/Photos from 2021/image000000.gif.json", "image000000.gif.json").
		addImage("Takeout/Google Photos/Photos from 2021/image000000.gif", 10).
		addImage("Takeout/Google Photos/Photos from 2021/image000000.jpg", 20)
}

func issue390WrongCount2() *inMemFS {
	return newInMemFS().
		addJSONImage("Takeout/Google Photos/2017 - Croatia/IMG_0170.jpg.json", "IMG_0170.jpg").
		addJSONImage("Takeout/Google Photos/Photos from 2018/IMG_0170.JPG.json", "IMG_0170.JPG").
		addJSONImage("Takeout/Google Photos/Photos from 2018/IMG_0170.HEIC.json", "IMG_0170.HEIC").
		addJSONImage("Takeout/Google Photos/Photos from 2023/IMG_0170.HEIC.json", "IMG_0170.HEIC").
		addJSONImage("Takeout/Google Photos/2018 - Cambodia/IMG_0170.JPG.json", "IMG_0170.JPG").
		addJSONImage("Takeout/Google Photos/2023 - Belize/IMG_0170.HEIC.json", "IMG_0170.HEIC").
		addJSONImage("Takeout/Google Photos/Photos from 2017/IMG_0170.jpg.json", "IMG_0170.jpg").
		addImage("Takeout/Google Photos/2017 - Croatia/IMG_0170.jpg", 514963).
		addImage("Takeout/Google Photos/Photos from 2018/IMG_0170.HEIC", 1332980).
		addImage("Takeout/Google Photos/Photos from 2018/IMG_0170.JPG", 4570661).
		addImage("Takeout/Google Photos/Photos from 2023/IMG_0170.MP4", 6024972).
		addImage("Takeout/Google Photos/Photos from 2023/IMG_0170.HEIC", 4443973).
		addImage("Takeout/Google Photos/Photos from 2018/IMG_0170.MP4", 2288647).
		addImage("Takeout/Google Photos/2018 - Cambodia/IMG_0170.JPG", 4570661).
		addImage("Takeout/Google Photos/2023 - Belize/IMG_0170.MP4", 6024972).
		addImage("Takeout/Google Photos/2023 - Belize/IMG_0170.HEIC", 4443973).
		addImage("Takeout/Google Photos/Photos from 2017/IMG_0170.jpg", 514963)
}
