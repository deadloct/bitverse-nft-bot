package internal

type BitVerseCollection struct {
	Singular string
	Name     string
	Address  string
}

var BitVerseCollections = map[string]BitVerseCollection{
	"hero": {
		Singular: "Hero",
		Name:     "BitVerse Heroes",
		Address:  "0x6465ef3009f3c474774f4afb607a5d600ea71d95",
	},
	"portal": {
		Singular: "Portal",
		Name:     "BitVerse Portals",
		Address:  "0xe4ac52f4b4a721d1d0ad8c9c689df401c2db7291",
	},
}
