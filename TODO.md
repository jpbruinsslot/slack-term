Bugs:

- [x] when switching channels sometimes messages are persisted in the new
      channel, the Buffer() in Chat will probably not go further than the
      latest message. Could be that items are added to List and not cleared
      when switching channels
- [x] send message as user, now it will send it as a bot    
- [x] alot of usernames 'unknown' should be a better way to uncover this
- [x] message creation in input.go and events.go should be made into function
      CreateMessage
- [x] restarting the application will always add the latest sent message
      through RTM in the selected channel
- [x] uncovering usernames takes too long, should find a better way
      test without uncovering, if that is the problem
- [ ] GetMessages for a channel can result in `json: cannot unmarshal number
      into Go value of type string` https://github.com/nlopes/slack/issues/92
- [ ] docs at exported functions

Features:

- [x] channel name in chat pane
- [x] new message indicator
- [ ] group channels, im channels
- [ ] scrolling in chat pane
- [ ] scrolling in channel pane
