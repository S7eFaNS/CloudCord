
import React, { useState } from 'react';
import './Chatpage.css';
import { useAuth0 } from '@auth0/auth0-react';
import Sidebar from '../Sidebar/Sidebar';

const ChatPage = ({ receiverName = "User123" }) => {
  const { user } = useAuth0();
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');

  const handleSend = () => {
    if (!input.trim()) return;

    const newMessage = {
      sender: user.nickname,
      content: input.trim(),
    };

    setMessages([...messages, newMessage]);
    setInput('');
  };

  return (
<div className="chatPage">
      <Sidebar />
      <div className="chatMain">
        <div className="chatBody">
          {messages.map((msg, index) => (
            <div
              key={index}
              className={msg.sender === user.nickname ? "myMessage" : "otherMessage"}
            >
              {msg.content}
            </div>
          ))}
        </div>

        <div className="chatInput">
          <input
            type="text"
            placeholder="Type a message..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleSend()}
          />
          <button onClick={handleSend}>Send</button>
        </div>
      </div>
    </div>
  );
};

export default ChatPage;