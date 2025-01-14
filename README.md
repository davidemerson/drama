# DRAMA FINDS YOU
This is a simple go program which will bestow the gift of drama upon neighboring open print queues. It will:

1. Scan your network (entered in CIDR) for IPP/AirPrint/RAW/LPD open hosts.
2. Give you a list of discovered open print queues.
3. Upon request, dend a PDF from its local directory at random to each discovered print queue.

The PDFs are intended to be recovered by whomever next passes the printer, and (hopefully) read aloud, dramatically. A spontaneous play nobody wanted but everyone will appreciate.

You need to install CUPS for this to work. Does best on *nix / MacOS as a result, but I suppose Windows Subsystem for Linux might work too.

`go build drama.go` gets you an executable in most cases. Place the executable in a directory with the two PDFs above, or ones of your own creation.

Feel free to submit additional dramas to this repo and they'll be considered for inclusion. They should be short and good reading.

You should not execute this on networks you don't own or have permission to use for this purpose.
