


export { 
    magenta = color { r = 0xAA, g = 0x00, b = 0xAA },
    yellow = color { r = 0xFF, g = 0xFF, b = 0x55 },
    white = color { r = 0xFF, g = 0xFF, b = 0xFF },
}

export {
    guybrush = actor {
        name = "guybrush",
        costume = ref("resources:costumes/Guybrush"),
        talkcolor = white,
    },
}

export {
    music1 = music("resources:audio/OnTheHill"),
    music2 = music("resources:audio/GuitarNoodling"),
    cricket = sound("resources:audio/Cricket"),
}

DEFAULT = defaults()

function DEFAULT.pickup()
    guybrush:say("I can't pick that up.")
end

function DEFAULT.use()
    guybrush:say("I can't use that.")
end

function DEFAULT.open()
    guybrush:say("I can't open that.")
end

function DEFAULT.close()
    guybrush:say("I can't close that.")
end

function DEFAULT.pull()
    guybrush:say("I can't pull that.")
end

function DEFAULT.push()
    guybrush:say("I can't push that.")
end

function DEFAULT.talkto(what)
    if what.type == "actor" then
        guybrush:say("It's not time for a chat.")
    else
        guybrush:say("I can't talk to that.")
    end
end

function DEFAULT.lookat()
    guybrush:say("There is nothing special about that.")
end

function DEFAULT.turnon()
    guybrush:say("I can't turn that on.")
end

function DEFAULT.turnoff()
    guybrush:say("I can't turn that off.")
end

function DEFAULT.give()
    guybrush:say("I can't give that.")
end

function DEFAULT.walkto()    
end

