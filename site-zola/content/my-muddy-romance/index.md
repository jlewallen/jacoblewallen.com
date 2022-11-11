+++
title = "My MUDdy Romance"
date = 2022-11-09T07:12:10-08:00
draft = false
authors = ["jacob"]
[taxonomies]
tags = ["MUDs", "history", "programming"]
+++

MUDs hold a very special place in my heart. If you aren't familiar, a
[MUD](https://en.wikipedia.org/wiki/MUD) is a multi-user, networked, text based
virtual world. Navigable via english like commands, `go north`, very nearly
like moving around the file system in the terminal. Hundreds of interconnected
rooms and areas forming whole towns, forests, caves, dungeons _and their
contents_. Entire land masses, all cohesively and fluidly described along with
the game specific mechanics and physical objects and characters of the virtual
world. They are amazing monuments to creativity, capable of setting your
imagination free and drawing you in. [^1]

Unfortunately, my actual experience with them is pretty limited and all begins
with GemStone III.

<center style="background: #eacd97; height: 150px; padding: 20px;">
  {{ img(src="gemstone_iii_logo.gif" alt="Logo for GemStone III") }}
</center>

> This is the heart of the main square of Wehnimer’s Landing. The impromptu shops
> of the bazaar are clustered around this central gathering place, where townfolk,
> travellers, and adventurers meet to talk, conspire or raise expeditions to the
> far-flung reaches of Elanthia. At the north end of the space, an old well, with
> moss-covered stones and a craggy roof, is shaded by a strong, robust tree. The
> oak is tall and straight, and it is apparent that the roots run deep. You also
> see a copper lockpick, a heavy backpack, a large acorn and some stone benches.

GS3 is a text based fantasy RPG and it was my first and only experience with
anything close to a MUD. My family was on
[Prodigy](https://en.wikipedia.org/wiki/Prodigy_(online_service)), and back
then all the big services had access to GS3. I wish I could remember how I was
introduced to the game. I just remember being obsessed. Thoroughly invested.
Staying up late, getting up early and throwing mini LAN parties with my other
friend who played. Printed, physical maps. Possibly laminated. I think you
understand. [^2]
 
Similarly, I don't remember why I stopped playing.

Anyway.

Lately, my side project has been writing a modern MUD engine, of sorts. It's
written in Python and is called [dimsum](https://github.com/jlewallen/dimsum/).
Its focus is on delivering a "rich-text" experience over the web, with the
ability to fall back on something more traditional. It's been a great project
and I'm proud of the work in that repository.

<div style="background: #020221;">
  {{ img(src="dimsum_logs.png" alt="Screenshot of verbose log output from dimsum which would be very difficult to understand without knowing more about how dimsum works.") }}
</div>

Which brings me to the purpose of this post.

My first real rust project is a port of this engine! :smile: In addition to
bringing along that feel-good, new project motivation it also feels like a
_smart_ project. I'm able to build on top of the architectural lessons I
learned with the Python implementation from the beginning and improve many
things along the way. It also fits my learning style, which typically involves
building a hobby project that in the past, nobody would have seen.

To begin with, it's effectively a language (simple English, to begin with)
parser and interpreter for carrying out actions in the virtual world. It
requires some interesting borrowing semantics across the domain model and the
service layer above where multiple domain objects interact. On top of that is
either a Web/RPC API or interactive shell. I also have several features in mind
for a kind of federation and linking of "domains" (effectively servers) and the
sharing of dynamically created game objects. Oh, and written in the most
self-extensible way possible.

Ideally, this is a world that can be used to build itself.

```sh
> look

"A lush, dense forest. Large pine trees of varrying thickness obscure the view."

> dig north to "A tiny meadow, bordering a pale lake."

"Whoa! What was that!"
~~~~
"A tiny meadow, bordering a pale lake."

> go south

"A lush, dense forest. Large pine trees of varrying thickness obscure the view."

```

It's a game that is it's own editor. Not entirely new, actually. There is a
deep history of this outside of MUDs/MOOs. Personally, I can only think of LISP
REPLs, though I'm sure there are so many more and even earlier examples. At any
rate, this doesn't mean that the entire world and it's mechanics need to be
written the same way.  I'm comfortable with a mixed language approach. I'll be
thinking about this as I work on the basics in rust.

```rust
#[test]
fn it_goes_through_routes() -> Result<()> {
    let args: ActionArgs = BuildActionArgs::new()?
        .ground(vec![QuickThing::Route(
            "East".to_string(),
            Box::new(QuickThing::Place("Place".to_string())),
        )])
        .try_into()?;

    let action = GoAction {
        item: Item::Route("east".to_string()),
    };
    let reply = action.perform(args.clone())?;
    let (_, person, area, _) = args.clone();

    assert_eq!(reply.to_json()?, SimpleReply::Done.to_json()?);

    assert_ne!(tools::area_of(&person)?.key(), area.key());

    Ok(())
}
```

I'm going to hold back on a URL, for now. At least until I've settled on a name
and have ironed out some organizational things, oh and an easier approach to
tests, haha! I would love to have a post describing the architecture, so
that'll be the next one I think.

Thanks for stopping by!


[^1] There are also [MOO](https://en.wikipedia.org/wiki/MOO)s, which are much
more like the kinds of things I'm interested in. I believe object-orientated
patterns are a great fit here. Sure, there are other patterns you could use.
Actor based approaches seem powerful to me, for example but they can usually be
thought about in an object-orientated way.

[^2] It's hard for me to tell how popular GemStone III still is. I think people
may still be playing, or its successor. I'm happy to leave my memories of the
game in the past.
You can still find artifacts about it online, like the
[Welcome to the Unofficial GemStone III Page](http://www.tamcon.com/GS3/) page,
or this really great stroll through time, simply titled
[GemStone III](https://www.angelfire.com/ca2/EtrionEmpire/gs3.html).
Oh and this is a great article about the game:
[Don’t Cry for Me, Elanthia: An Archaeology of Gemstone III](https://theappendix.net/issues/2014/10/dont-cry-for-me-elanthia-an-archaeology-of-gemstone-iii)
which I really, really enjoyed reading and only found while researching this post!
