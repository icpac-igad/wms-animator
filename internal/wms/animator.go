package wms

import (
	"bytes"
	"image"
	"image/gif"
)

func GenerateGif(images []image.Image, fps int) (*bytes.Buffer, error) {
	im_p := encodeImgPaletted(&images)

	g := &gif.GIF{}

	delay := 100 / fps

	for _, i := range im_p {
		g.Image = append(g.Image, i)
		g.Delay = append(g.Delay, delay)
	}

	buf := new(bytes.Buffer)
	err := gif.EncodeAll(buf, g)

	if err != nil {
		return nil, err
	}

	return buf, nil
}

func encodeImgPaletted(images *[]image.Image) []*image.Paletted {
	// Gif options
	opt := gif.Options{}
	g := []*image.Paletted{}

	for _, im := range *images {
		b := bytes.Buffer{}
		// Write img2gif file to buffer.
		err := gif.Encode(&b, im, &opt)

		if err != nil {
			println(err)
		}
		// Decode img2gif file from buffer to img.
		img, err := gif.Decode(&b)

		if err != nil {
			println(err)
		}

		// Cast img.
		i, ok := img.(*image.Paletted)
		if ok {
			g = append(g, i)
		}
	}
	return g
}
