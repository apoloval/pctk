decl = import("resources:scripts/melee.decl")
common = import("resources:scripts/common")

melee = decl.melee
guybrush = common.guybrush

pirates = actor {
    name = "men of low moral fiber (pirates)",
    size = size {w=60, h=64},
    talkcolor = common.magenta,
    usepos = pos {x=90, y=128},
    usedir = LEFT
}

function melee:enter()
    local skipintro = true
    melee.box4:disable()

    pirates:show {
        pos = pos {x=38, y=137},         
        lookat = RIGHT,
    }

    guybrush:show {
        pos = pos {x=340, y=140}, 
        lookat = LEFT,
    }
    
    common.music1:play()
    common.cricket:play()
    guybrush:walkto(pos {x=290, y=140}):wait()
    if not skipintro then
        CONTROL:paneoff()
        CONTROL:cursoroff()

        guybrush:say("Hello, I'm Guybrush Threepwood,\nmighty pirate!"):wait()
        pirates:say("**Oh no! This guy again!**")
        guybrush:walkto(pos{x=120, y=140}):wait()
        guybrush:say("I think I've lost the keys to my boat."):wait()
        guybrush:say("Have you seen any keys?"):wait()
        sleep(1000)
        pirates:say("Eeerrrr... Nope!"):wait()
        sleep(1000)
        
        common.music2:play()
        guybrush:walkto(pos{x=120, y=120}):wait()
        guybrush:say("Where can I find the keys?"):wait()
        sleep(1000)
        guybrush:walkto(pos{x=120, y=140}):wait()
        guybrush:say("Ooooook..."):wait()
        sleep(2000)
        guybrush:lookdir(RIGHT)
        sleep(2000)
        guybrush:say("Ok, I will try the Scumm bar."):wait()
        guybrush:lookdir(LEFT)
        guybrush:say("Thank you guys!"):wait()
        common.cricket:play()
        guybrush:walkto(pos{x=360, y=140}):wait()
        sleep(1000)
        
        pirates:say("Oh, Jesus! I though he would\ntell again that stupid\ntale about LeChuck!"):wait()
        sleep(5000)
        pirates:say("Who has the keys?", { color = common.yellow }):wait()
        sleep(1000)
        pirates:say("Me!")
        sleep(2000)
    end

    guybrush:select()
    CONTROL:paneon()
    CONTROL:cursoron()
    self:camfollow(guybrush)
end

function melee.bucket:walkto()
    guybrush:say("I will not get closer")
end

function melee.bucket:lookat()
    if self.owner == guybrush then
        guybrush:say("It's a empty bucket.\nBut it's ALL MINE!")
    else
        guybrush:say("It's a empty bucket.")
    end
end

function melee.bucket:give(to)
    if to == pirates then
        guybrush:say("I'd rather not. I am afraid\nthey'd get attached to it.")
    else
        DEFAULT.give(self)
    end
end

function melee.bucket:pickup()
    CONTROL:cursoroff()
    guybrush:say("I don't know how this could help\nme to find the keys, but..."):wait()
    guybrush:toinventory(self)
    melee.box3:enable()
    melee.box4:enable()
    CONTROL:cursoron()
end

function melee.bucket:use(on)
    if on == melee.clock then
        guybrush:say("Time flies, but I don't think\nI can gather it in the bucket.")
    elseif on == pirates then
        melee.bucket:give(pirates)
    else
        DEFAULT.use(self, on)
    end
end

function melee.clock:lookat()
    guybrush:say("It's weird. I have the feeling\nthat the time is not passing."):wait()
end

function melee.clock:turnon()
    guybrush:say("Do I look like a watchmaker?"):wait()
end

function melee.clock:turnoff()
    guybrush:say("Well, I guess I couldn't be more off"):wait()
end

function pirates:lookat() 
    guybrush:say("They didn't move since I arrived\nin Monkey Island I."):wait()
    guybrush:say("I guess they are waiting for\nsomething..."):wait()
end

function pirates:talkto()
    while true do
        local choice = CONTROL:sentencechoice()
        choice:add("Hello, guys! Are you really\nsure you don't have the keys?")
        choice:add("I'm sure you have them! I know you!")
        choice:add("I'm Guybrush Threepwood, mighty pirate!")
        choice:add("I think I will go to the Scumm bar. See you!")
        local s = choice:waitsay()
        if s == 1 then
            pirates:say("Yes, we are sure!"):wait()
            pirates:say(
                "If we had them, we\nwould have left\nthis place long ago!",
                { color = common.yellow }
            ):wait()
            pirates:say("For sure!"):wait()
        end
        if s == 2 then
            pirates:say("No, we don't have them!"):wait()
            pirates:say("And no, you don't know us!", { color = common.yellow }):wait()
        end
        if s == 3 then
            pirates:say("Yes, we know it! You killed LeChuck"):wait()
            pirates:say("You are a real hero", { color = common.yellow }):wait()
            pirates:say("**Oh, Jesus!**"):wait()
        end
        if s == 4 then
            CONTROL:paneon()
            return
        end
    end
end
