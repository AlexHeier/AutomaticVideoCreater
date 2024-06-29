package global

/**
VAR THAT CAN BE CHANGED
*/

// Youtube
var YoutubeChannelName string = "QuotePixel"

var PostYoutubeVideo bool = true
var DeleteYoutubeVideoAfterPost bool = true
var MadeForKids bool = false

// TikTok
var TikTokChannelName string = "TheRedditPixel"

var PostTikTokVideo bool = false
var DeleteTikTokVideoAfterPost bool = false

// AI voice
var VoiceID string = "Liv"    // Other voices: Scarlett, Liv, Amy, Dan and Will
var VoiceSpeed string = "0.1" // -1 to 1
var VoicePitch string = "1"   // 0.5 to 1.5
var Bitrate string = "192k"   // 320k, 256k, 192k, ...

// Text on screen
var BorderThickness int = 10

// max 15 entries
// make sure to test the new thema to make sure there exsist a quote and video for said thema.
var Themas []string = []string{
	"love",
	"friendship",
	"happiness",
	"life",
	"courage",
	"trust",
}

/**
DO NOT CHANGE
*/

// The Youtube # limit
const TagLimit int = 15

const MaxVoiceCharacters int = 1500
