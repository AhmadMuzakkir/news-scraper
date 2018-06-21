package model

const (
	NstName = "New Straits Times"
	NstId   = "nst"

	BharianName = "Berita Harian"
	BharianId   = "bharian"

	UtusanName = "Utusan"
	UtusanId   = "utusan"
)

var BhSources = []NewsSource{
	NewNewsSourceBh("Berita", "Nasional", "https://www.bharian.com.my/berita/nasional", []string{"news", "nation"}),
	NewNewsSourceBh("Berita", "Politik", "https://www.bharian.com.my/berita/politik", []string{"news", "politics"}),
	NewNewsSourceBh("Berita", "Kes", "https://www.bharian.com.my/berita/kes", []string{"news", "crime"}),
}

var NstSources = []NewsSource{
	NewNewsSourceNst("News", "Nation", "https://www.nst.com.my/news/nation", []string{"news", "nation"}),
	NewNewsSourceNst("News", "Politics", "https://www.nst.com.my/news/politics", []string{"news", "politics"}),
	NewNewsSourceNst("News", "Crime & Courts", "https://www.nst.com.my/news/crime-courts", []string{"news", "crime"}),
	NewNewsSourceNst("News", "Exclusive", "https://www.nst.com.my/news/exclusive", []string{"news", "exclusive"}),
	NewNewsSourceNst("News", "Government", "https://www.nst.com.my/news/government-public-policy", []string{"news", "government"}),
}

var UtusanSources = []NewsSource{
	NewNewsSourceUtusan("Berita", "Terkini", "http://www.utusan.com.my/berita/terkini", []string{"news", "latest"}),
	NewNewsSourceUtusan("Berita", "Utama", "http://www.utusan.com.my/berita/utama", []string{"news", "main"}),
	NewNewsSourceUtusan("Berita", "Nasional", "http://www.utusan.com.my/berita/nasional", []string{"news", "nation"}),
	NewNewsSourceUtusan("Berita", "Politik", "http://www.utusan.com.my/berita/politik", []string{"news", "politics"}),
	NewNewsSourceUtusan("Berita", "Jenayah", "http://www.utusan.com.my/berita/jenayah", []string{"news", "crime"}),
}

type NewsSource struct {
	NewspaperName       string   `json:"name"`
	NewspaperId         string   `json:"id"`
	OriginalCategory    string   `json:"category"`
	OriginalSubcategory string   `json:"subcategory"`
	Tags                []string `json:"tags"`
	Url                 string   `json:"url"`
}

func NewNewsSourceNst(originalCategory, originalSubcategory, url string, tags []string) NewsSource {
	ns := NewsSource{
		NewspaperName:       NstName,
		NewspaperId:         NstId,
		OriginalCategory:    originalCategory,
		OriginalSubcategory: originalSubcategory,
		Url:                 url,
		Tags:                tags,
	}

	return ns
}

func NewNewsSourceBh(originalCategory, originalSubcategory, url string, tags []string) NewsSource {
	return NewsSource{
		NewspaperName:       BharianName,
		NewspaperId:         BharianId,
		OriginalCategory:    originalCategory,
		OriginalSubcategory: originalSubcategory,
		Url:                 url,
		Tags:                tags,
	}
}

func NewNewsSourceUtusan(originalCategory, originalSubcategory, url string, tags []string) NewsSource {
	return NewsSource{
		NewspaperName:       UtusanName,
		NewspaperId:         UtusanId,
		OriginalCategory:    originalCategory,
		OriginalSubcategory: originalSubcategory,
		Url:                 url,
		Tags:                tags,
	}
}
