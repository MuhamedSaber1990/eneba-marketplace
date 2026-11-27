# Cover and banner art: tool, date and prompts

The 24 game covers and the 3 homepage banners are AI-generated. This file
records how, so anyone can see what was asked for each image.

Tool: OpenAI Codex CLI 0.172.0 (`codex exec`), signed in with a ChatGPT
account, using its built-in image generation. The CLI does not report which
image model it used.

When: 2026-11-26. One run per image.

The files in this repository are the converted ones:
`frontend/public/covers/` (24 covers, 450 x 600 WebP) and
`frontend/public/banners/` (3 banners, 1600 x 640 WebP). They were made from
the originals with ffmpeg and libwebp: covers scaled and cropped at quality
78, banners cropped to their middle band (2.5 to 1) at quality 76. The
originals (covers 1086 x 1448, banners 1536 x 1024) and the rejected images
are kept outside the repository.

Each image was looked at before it was accepted, for text, logos and
likeness to existing games or characters. That check can catch a likeness to
something well known. It cannot rule out a likeness to work the checkers did
not know.

## How the set was made

1. The first cover, Thunder Circuit Kart, was made alone as a test.
2. The other 26 images were then made with that cover attached as a style
   reference. Their prompts include a sentence asking to match the
   reference's rendering but not its subject.
3. Nine images were rejected and made again, without the reference image.
   The reference had leaked its character into four covers. The reason is noted on each
   entry below.

The prompts below are the ones that produced the final images.

## thunder-circuit-kart (the first image, made without a reference)

```
Use your image generation tool to create ONE image and save it as the file originals/thunder-circuit-kart.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a single chunky go-kart with an original helmeted cartoon driver, drifting around a rain-soaked race track corner while a lightning bolt splits the stormy sky behind it.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color with a magenta-to-orange palette, clean shapes, slight cel-shading, no photorealism.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## ashen-crown

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/ashen-crown.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: an armored knight seen from behind, raising a sword toward a burning fortress on a hill.
Palette: ember orange and charcoal with gold highlights.
Mood: heroic and dramatic.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## starfall-odyssey

Made again. The first version was rejected: the pilot was a copy of the kart driver from the reference image.

```
Use your image generation tool to create ONE image and save it as the file originals/starfall-odyssey.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a small patched-together spaceship seen from outside, with no pilot or person visible anywhere, flying toward a shattered glowing planet while stars fall around it.
Palette: deep indigo and cyan with touches of gold.
Mood: full of wonder.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The finish is polished and detailed, like premium key art for a game store, with soft glow, crisp edges and visible brushwork.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## neon-drift-2089

Made again. The first version was rejected: the driver was a copy of the kart driver, and the palette was too close to the kart cover.

```
Use your image generation tool to create ONE image and save it as the file originals/neon-drift-2089.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a sleek futuristic street racing car with dark tinted windows and no driver or person visible, drifting through a rain-slick city street at night under glowing signs that carry no writing.
Palette: electric cyan and violet with acid yellow accents on near-black, no pink or magenta.
Mood: fast and energetic.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The finish is polished and detailed, like premium key art for a game store, with soft glow, crisp edges and visible brushwork.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## hollow-tide

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/hollow-tide.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a lone lighthouse on a rock while a dark sea rises around it, a tiny keeper with a lantern at its base, the beam cutting through fog.
Palette: cold teal and near-black with one pale yellow light.
Mood: ominous and eerie, quiet horror, not cute, no cartoon faces.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## ironvale-tactics

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/ironvale-tactics.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a stern commander in a long coat planting a plain banner on a snowy mountain ridge, rows of tiny soldiers in the valley below.
Palette: steel blue and slate with muted red accents.
Mood: serious and strategic.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## pocket-kingdoms

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/pocket-kingdoms.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a tiny cheerful castle town on a round green hilltop, like a diorama, with little flags and a windmill.
Palette: fresh green, sky blue and warm yellow.
Mood: cozy and playful.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## skybound-couriers

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/skybound-couriers.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a patched-up seaplane loaded with parcels flying between floating islands with waterfalls.
Palette: sky blue and cloud white with sunny orange accents.
Mood: breezy and optimistic.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## emberwatch

Made again. The first version was rejected: a twin of Ashen Crown: a figure from behind against fire, in the same palette.

```
Use your image generation tool to create ONE image and save it as the file originals/emberwatch.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: three rescue fighters in heavy protective gear with closed visors, faces hidden, walking toward the viewer down a city street at night, seen from the front, fires glowing in the windows of the buildings behind them and embers in the air.
Palette: night blues and deep navy, with the fire as a small orange accent.
Mood: tense and urgent.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The finish is polished and detailed, like premium key art for a game store, with soft glow, crisp edges and visible brushwork.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## lanternwood

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/lanternwood.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a small cottage and an apple tree at the edge of an old forest at dusk, paper lanterns glowing among the branches.
Palette: warm amber, moss green and twilight purple.
Mood: cozy and gently magical.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## circuit-breakers

Made again. The first version was rejected: the blue robot with a face screen was too close to an existing game character.

```
Use your image generation tool to create ONE image and save it as the file originals/circuit-breakers.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: two small boxy robots built from square riveted metal panels, each with one round camera lens for an eye and a short antenna, no face screens and no smiling faces, one mustard yellow and one cobalt blue, leaping in mid-air and joined by a glowing cable, inside a factory of platforms.
Palette: mustard yellow and cobalt blue on dark navy.
Mood: playful and precise.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The finish is polished and detailed, like premium key art for a game store, with soft glow, crisp edges and visible brushwork.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## frostline-rally

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/frostline-rally.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a rally car throwing up snow as it slides through a frozen mountain pass between pine trees.
Palette: ice blue and white with red accents on the car.
Mood: crisp and fast.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## dune-runners

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/dune-runners.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a wind-powered sand skiff with a tall sail racing across dunes toward distant ruins at dusk.
Palette: amber and burnt orange under a dusky violet sky.
Mood: adventurous.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## mythforge-legends

Made again. The first version was rejected: the blacksmith had the kart driver's hair and face.

```
Use your image generation tool to create ONE image and save it as the file originals/mythforge-legends.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a broad-shouldered older blacksmith with a short grey beard, cropped grey hair and a leather apron, hammering a glowing sword on an anvil while the shadow of a horned beast looms behind.
Palette: molten gold, crimson and deep bronze.
Mood: epic.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The finish is polished and detailed, like premium key art for a game store, with soft glow, crisp edges and visible brushwork.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## quiet-harbor

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/quiet-harbor.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a small fishing boat moored at the dock of a pastel fishing town at sunrise, gulls overhead.
Palette: soft peach, seafoam and pale blue.
Mood: calm and gentle.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## bladeborn-arena

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/bladeborn-arena.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a lone duelist in a low fighting stance with a curved blade, sparks flying, the arena crowd only blurred lights.
Palette: crimson and steel grey with white highlights.
Mood: fierce and focused.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## verdant-protocol

Made again. The first version was rejected: the figure had the kart driver's hair and colors.

```
Use your image generation tool to create ONE image and save it as the file originals/verdant-protocol.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a lone figure in a bulky white and grey sealed sci-fi suit with a fully opaque mirrored visor, no face and no hair visible, standing inside a gigantic overgrown greenhouse on a space station with huge glowing vines everywhere.
Palette: toxic green and teal on dark.
Mood: tense science fiction.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The finish is polished and detailed, like premium key art for a game store, with soft glow, crisp edges and visible brushwork.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## tiny-titans-league

Made again. The first version was rejected: a caricature that did not match the rendering of the other covers.

```
Use your image generation tool to create ONE image and save it as the file originals/tiny-titans-league.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a young footballer with short black hair and a calm determined look, drawn with a slightly oversized head like a stylized collectible figure, striking a ball in a small stadium, plain orange kit with no badge, painterly rendering, not a caricature, no exaggerated grin, no tongue.
Palette: grass green, bright white and sunny yellow with an orange kit.
Mood: fun and sporty.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The finish is polished and detailed, like premium key art for a game store, with soft glow, crisp edges and visible brushwork.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## obsidian-vow

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/obsidian-vow.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a hooded figure crouched on a rooftop, seen as a silhouette, in a city built inside a volcano crater with lava glowing far below.
Palette: obsidian black and deep purple with a thin line of molten orange.
Mood: dark, quiet and menacing, not cute, no cartoon faces.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## gearheart-garage

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/gearheart-garage.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a classic car body raised on a lift in a cluttered garage, a mechanic with a wrench beside it under a warm work lamp.
Palette: rust orange, teal and cream.
Mood: warm and hands-on.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## witchlight-academy

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/witchlight-academy.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a young apprentice witch in a pointed hat holding up a glowing lantern in front of a towering crooked castle at night, no glasses, no scarf, no owl.
Palette: violet and midnight blue with an emerald glow.
Mood: mysterious and inviting.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## rogue-meridian

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/rogue-meridian.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a steam train speeding through a stormy wasteland, a caped figure standing on a carriage roof with a fan of glowing blank cards.
Palette: maroon, brass and storm grey.
Mood: gritty and restless.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## saltwind-pirates

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/saltwind-pirates.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a sailing ship with patched sails cutting through turquoise waves past tropical islands, a captain at the bow, plain sails and flags with no emblem.
Palette: turquoise, sandy yellow and coral red.
Mood: swashbuckling and bright.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## paper-planets

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/paper-planets.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: stylized video game key art, portrait orientation, 3:4 aspect ratio (for example 1024x1536 or the closest portrait size available).
Subject: a tiny planet made of folded origami paper with paper trees and paper houses, floating among paper stars.
Palette: cream, pastel coral and mint on soft navy.
Mood: gentle and clever.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the subject sits in the upper two thirds; the bottom third is darker and simpler, because a title will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## banner-keys

Made again. The first version was rejected: the key and chest filled the whole height, so the wide crop cut both.

```
Use your image generation tool to create ONE image and save it as the file originals/banner-keys.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: a wide banner illustration in the style of video game key art, landscape orientation, 3:2 aspect ratio (for example 1536x1024 or the closest landscape size available).
Subject: a large ornate golden key floating and glowing above an open treasure chest that spills light, small simple shapes of a sword, a planet and a racing car drifting in the glow.
Palette: royal purple and gold.
Mood: inviting and a little magical.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The finish is polished and detailed, like premium key art for a game store, with soft glow, crisp edges and visible brushwork.
Composition: the image will be cropped to a very wide strip, so the whole subject must be small enough to fit inside the middle half of the image height, with plain empty background above it and below it. Put the subject on the right half; keep the left half darker and simpler, because text will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## banner-new

Made with the reference image attached (see above).

```
Use your image generation tool to create ONE image and save it as the file originals/banner-new.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: a wide banner illustration in the style of video game key art, landscape orientation, 3:2 aspect ratio (for example 1536x1024 or the closest landscape size available).
Subject: a bright comet streaking across the sky over a stylized landscape of hills at dawn, the feeling of something new arriving.
Palette: magenta fading into deep blue with a warm glow on the horizon.
Mood: fresh and hopeful.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The attached image is a cover from the same shop: match its rendering technique, level of detail and finish, but NOT its subject, characters or colors.
Composition: the image will be cropped to a very wide strip, so keep everything important inside the middle horizontal band. Put the subject on the right half; keep the left half darker and simpler, because text will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

## banner-play

Made again. The first version was rejected: the creatures were too close to an existing franchise.

```
Use your image generation tool to create ONE image and save it as the file originals/banner-play.png in the current directory. Do not write any code or other files, and do not modify anything else.

The image: a wide banner illustration in the style of video game key art, landscape orientation, 3:2 aspect ratio (for example 1536x1024 or the closest landscape size available).
Subject: colorful toy-like vehicles, a little rocket, a hot-air balloon and a small buggy, flying and bouncing over a floating playground of platforms, slides and bunches of balloons, with no creatures, no animals, no people, no game consoles and no controllers anywhere.
Palette: warm orange and red with touches of yellow.
Mood: lively and friendly.
Style: bold painterly digital illustration, one strong central subject, dramatic lighting, rich saturated color, clean shapes, slight cel-shading, no photorealism. The finish is polished and detailed, like premium key art for a game store, with soft glow, crisp edges and visible brushwork.
Composition: the image will be cropped to a very wide strip, so the whole subject must be small enough to fit inside the middle half of the image height, with plain empty background above it and below it. Put the subject on the right half; keep the left half darker and simpler, because text will be overlaid there later.
Strictly: no text, no letters, no numbers, no logos, no emblems, no watermarks, no user interface, no real brands, no existing game or film characters, nothing resembling an existing franchise.

When done, reply with only the saved file path and the pixel size of the image.
```

