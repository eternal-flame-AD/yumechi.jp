#!/usr/bin/env fish

set -l convert_files "icon_do_not_connect.gif"

for icon in icon_*.gif
    set -l outfile (echo $icon | sed 's/icon_/icon_t_/g')


        magick $icon -bordercolor white -border 1x1 \
            -alpha set -channel RGBA -fuzz 5% \
            -fill none -floodfill +0+0 white \
            -shave 1x1 $outfile

end

