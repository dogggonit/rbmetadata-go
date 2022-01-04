# Translation
|  filename  | Approx. Lines |                            commments                           |
|------------|---------------|----------------------------------------------------------------|
|aac.c       |  90           | looks complicated                                              |
|aiff.c      |  91           | looping, looks complicated                                     |
|spc.c       |  93           | iso decode, video game                                         |
|ay.c        |  94           | lots of renaming, video game                                   |
|wavpack.c   | 101           | looping, doesn't look bad                                      |
|ape.c       | 110           | album art, cue sheet, long                                     |
|vtx.c       | 111           | doesn't look bad, video game                                   |
|vgm.c       | 119           | not terrible looking, video game                               |
|ogg.c       | 119           | seems sparse, requires vorbis                                  |
|mpc.c       | 134           | requires replaygain                                            |
|oma.c       | 136           | weird proprietary?, from ffmpeg, macros, not terrible looking  |
|nsf.c       | 167           | video game, mostly cases, doesn't look bad                     |
|asap.c      | 174           | pointers, mostly parsing, doesn't look that bad                |
|wave.c      | 286           | looping, lot going on here                                     |
|smaf.c      | 286           | looks complicated, lots of prints                              |
|asf.c       | 381           | requires replaygain                                            |
|rm.c        | 392           | lots of defines and prints, looks a little complicated         |

# Project
- [ ] Finish writing parsers
- [ ] Solidify library API to make use with other go code easier
- [ ] Write tests