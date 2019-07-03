VERSION = "0.0.2"

function onViewOpen(view)

    local ft = view.Buf.Settings["filetype"]
    
    if  ft == "go" or
        ft == "perl" then
        SetLocalOption("tabstospaces", "false", view)
        SetLocalOption("smartindent", "true", view)
    elseif ft == "xslt" or
        ft == "html" or
        ft == "makefile" then
        SetLocalOption("tabstospaces", "false", view)
        SetLocalOption("smartindent", "false", view)
    elseif ft == "fish" or
           ft == "python" or
           ft == "python2" or
           ft == "python3" or
           ft == "yaml" or
           ft == "nim" then
        SetLocalOption("tabstospaces", "true", view)
        SetLocalOption("smartindent", "false", view)
    end
end
