export {
    melee = room {
        background = ref("resources:backgrounds/Melee"),
        walkboxes = {
            box0 = walkbox {
                vertices = {
                    pos {x=5, y=140}, 
                    pos {x=475, y=140}, 
                    pos {x=475, y=117}, 
                    pos {x=80, y=117},
                }, 
                scale = 1,
            },
            box1 = walkbox {
                vertices = {
                    pos {x=80, y=117}, 
                    pos {x=300, y=117}, 
                    pos {x=230, y=100}, 
                    pos {x=100, y=100},
                }, 
                scale = 0.95,
            },
	        box2 = walkbox {
                vertices = { 
                    pos {x=100, y=100}, 
                    pos {x=230, y=100}, 
                    pos {x=215, y=90}, 
                    pos {x=115, y=90}
                }, 
                scale = 0.8,
            },
	        box3 = walkbox {
                vertices = { 
                    pos {x=115, y=90}, 
                    pos {x=215, y=90}, 
                    pos {x=197, y=82}, 
                    pos {x=128, y=82},
                }, 
                scale = 0.6,
                enabled = false,
            },
	        box4 = walkbox {
                vertices = {
                    pos {x=155, y=82}, 
                    pos {x=165, y=82}, 
                    pos {x=165, y=75}, 
                    pos {x=155, y=75},
                }, 
                scale = 0.3,
            },
        },
        bucket = object {
            class = APPLICABLE,
            name = "bucket",
            sprites = ref("resources:sprites/objects"),
            pos = pos {x=440, y=130},
            hotspot = rect {x=430, y=110, w=20, h=20},
            usedir = RIGHT,
            usepos = pos {x=420, y=130},
            default = state {
                anim = sequenceanim { 
                    duration = 100,
                    row = 6, 
                    columns = {5}},
            },
            pickup = state {},
            state = "default"
        },
        clock = object {
            name = "clock",
            hotspot = rect {x=150, y=25, w=24, h=18},
            usedir = UP,
            usepos = pos {x=161, y=116}
        },
    }
}

