#
# markov.py -
# markov chain bot that uses yamamura logfiles as input
#

# import the needed modules
from multiprocessing import Process, Queue
from discord.ext.commands import Bot
import re, markovify, discord, json

# load config file, exit if not found
cfg = None
try:
    with open("config.json", "r") as cfgFile:
        cfg = json.load(cfgFile)
except FileNotFoundError:
    print(
        'copy "config.example.json", rename it to "config.json and edit it before running this bot'
    )

# create the bot object
bot = Bot(description="that one markov chain bot", command_prefix="m~")

@bot.event
async def on_message(msg):
 if msg.author == bot.user:
    return
 if msg.author.bot:
    return
 await bot.process_commands(msg)

# markov chain
chain = None

# the spooler queue
spool_queue = None

# the thing that fixes mentions
class squeaky:

    # quick workaround for passing
    # context
    def __init__(self, context):

        # set the context
        self.ctx = context

    # whem this class is called
    def __call__(self, match):

        # really quick and dirty workaround
        try:

            # attempt to get length
            len(match)

        except:

            # group them
            match = match.group(0)

        # getting uid
        try:

            # get the uid
            uid = int(match[2:(len(match) - 1)])

        except:

            try:

                # some are like this
                uid = int(match[3:(len(match) - 1)])

            except:

                # otherwise, return
                return match

        try:

            # look it up
            return self.ctx.message.guild.get_member(uid).name

        except:

            # it might be a role
            for role in self.ctx.message.guild.roles:

                # check if it is here
                if role.id == uid:

                    # return the role name
                    return role.name

            # otherwise, return uid
            return str(uid)


# the callback that the spooler runs after each write
def spooler_callback(logfile):

    # reset the chain
    global chain
    try:

        # ensure no errors occur
        chain = markovify.NewlineText(logfile.read())

    except:

        pass


# the function we use to spool writes to the logfile
def spooler(queue, callback):

    # load the log file
    with open("chat.log", "a+") as f:

        # loop to constantly write lines to the file
        while True:

            # get a message to write from the queue
            message = queue.get()

			# ensure its not empty
            if message == "":

                continue

            # write the message
            f.write(message + "\n")

            # run the callback with the file
            callback(f)


@bot.event
async def on_ready():

    # print some output
    print(
        f"logged in as: { bot.user.name } id:{ bot.user.id } | connected to { str(len(bot.guilds)) } server(s)"
    )
    print(
        f"invite: https://discordapp.com/oauth2/authorize?bot_id={ bot.user.id }&scope=bot&permissions=8"
    )
    await bot.change_presence(activity=discord.Game(name="scanning your text..."))

    # make the chain

    # extracted text from file variable
    extracted_text = ""

    # load the log file
    with open("chat.log", "a+") as f:

        # extract all text from it
        extracted_text = f.read()

    # save the extracted text to a markov chain
    global chain
    try:

        # ensure no errors occur
        chain = markovify.NewlineText(extracted_text)

    except:

        pass

    # create the spooler queue
    global spool_queue
    spool_queue = Queue()

    # start up the spooler
    spooler_proc = Process(target=spooler, args=(spool_queue, spooler_callback))
    spooler_proc.start()


# output each new message to a file
@bot.event
async def on_message(message):

    # get the arguments passed
    args = message.content.split(" ")[1:]

    # check if this is a command
    if message.content.split(" ")[0] == (cfg["prefix"] + "markov"):

        # make the sentence variable
        sentence = None

        # get the chain
        global chain

        # check if we have arguments
        if args == []:

            # make a sentence
            while sentence == None:

                # make a sentence
                sentence = chain.make_sentence(tries=250)

        else:

            # make a variable for the
            # to-be-sent message
            sentence = ""

            # individual sentence var
            isentence = None

            # make a variable for the
            # number of sentences we
            # have constructed
            x = 1

            # check if the arg is a valid
            # int
            try:

                # check
                iters = int(args[0])

                # tell them if it is definitely too big
                if iters >= 50:

                    # send a message
                    await message.channel.send(
                        f"{ ctx.author.mention }, that number of sentences is definitely way too high!"
                    )

                    # exit
                    return

            except ValueError:

                # respond saying it isn't valid
                await message.channel.send(f"{ args[0] } is not a number...")

                # exit
                return

            # now construct the message
            for x in range(iters):

                # make sure we have a message
                while isentence == None:

                    # make the message
                    isentence = chain.make_sentence(tries=250)

                # if it works, append the message to the
                # full message
                sentence += ("%s\n" % isentence)

                # reset the isentence variable
                isentence = None

        # make the message squeaky clean
        sentence = re.sub(r"\<\@(.*?)\>", squeaky(ctx), sentence, flags=re.IGNORECASE)

        # remove the @everyones and @heres
        sentence = re.sub(r"\@everyone", "everyone", sentence, flags=re.IGNORECASE)
        sentence = re.sub(r"\@here", "here", sentence, flags=re.IGNORECASE)

        try:

            # send it
            await message.channel.send(sentence)

        except discord.errors.HTTPException:

            # send an error message
            await message.channel.send(
                f"{ ctx.author.mention }, sorry, but the message generated was too long..."
            )

    else:

        # get the message and let the spooler take care of it
        global spool_queue
        try:

            # ensure we don't kill it before it started
            spool_queue.put(message.clean_content)

        except:

            pass


# run the bot
bot.run(cfg["token"], bot=True)
