include("resources:scripts/common")

melee = room {
    background = "resources:backgrounds/Melee",
    objects = {
        bucket = {
            name = "bucket",
            sprites = "resources:sprites/objects",
            pos = {x=300, y=180},
            hotspot = {x=290, y=190, w=20, h=20},
            usedir = left,
            usepos = {x=290, y=190},
            states = {
                default = {
                    anim = { index = 0, frames = 0, delay = 0 },
                },
                pickup = {}
            }
        }
    }
}

function melee.enter(self)
    local pirate1_dialog_props = { pos = {x=60, y=20}, color = magenta }
    local pirate2_dialog_props = { pos = {x=60, y=50}, color = yellow }

    guybrush:show{
        pos={x=340, y=140}, 
        dir=left,
        costume="resources:costumes/Guybrush" -- TODO: this must be a variable
    }
    
    music1:play()
    cricket:play()
    guybrush:walkto({x=290, y=140}).wait()
    guybrush:say("Hello, I'm Guybrush Threepwood,\nmighty pirate!").wait()
    sayline("**Oh no! This guy again!**", pirate1_dialog_props)
    guybrush:walkto({x=120, y=140}).wait()
    guybrush:say("I think I've lost the keys to my boat.").wait()
    guybrush:say("Have you seen any keys?", {delay=2000}).wait()
    sayline("Eeerrrr... Nope!", pirate1_dialog_props)
    sleep(2000)
    
    music2:play()
    guybrush:walkto({x=120, y=120}).wait()
    guybrush:say("Where can I find the keys?", {delay=1000}).wait()
    guybrush:walkto({x=120, y=140}).wait()
    guybrush:say("Ooooook...").wait()
    sleep(2000)
    guybrush:stand({dir = right}).wait()
    sleep(2000)
    guybrush:say("Ok, I will try the Scumm bar.").wait()
    guybrush:stand({dir = left}).wait()
    guybrush:say("Thank you guys!").wait()
    cricket:play()
    guybrush:walkto({x=360, y=140}).wait()
    
    sayline("Oh, Jesus! I though he would\ntell again that stupid\ntale about LeChuck!", pirate1_dialog_props).wait()
    sleep(5000)
    sayline("Who has the keys?", pirate2_dialog_props).wait()
    sleep(1000)
    sayline("Me!", pirate1_dialog_props)
    
    guybrush:select()
    usercontrol(true)          
end