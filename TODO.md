Bugs:

- [ ] when switching channels sometimes messages are persisted in the new
      channel, the Buffer() in Chat will probably not go further than the
      latest message. Could be that items are added to List and not cleared
      when switching channels
- [ ] GetMessages for a channel can result in `json: cannot unmarshal number
      into Go value of type string` https://github.com/nlopes/slack/issues/92
- [ ] send message as user, now it will send it as a bot    
- [ ] alot of usernames 'unknown' should be a better way to uncover this
- [ ] uncovering usernames takes too long, should find a better way
- [ ] docs at exported functions
- [ ] message creation in input.go and events.go should be made into function
      CreateMessage
- [ ] restarting the application will always add the latest sent message
      through RTM in the selected channel

Features:

- [ ] scrolling in chat pane
- [ ] scrolling in channel pane
- [x] channel name in chat pane
