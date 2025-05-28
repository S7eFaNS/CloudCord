import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import './Chatpage.css';
import { useAuth0 } from '@auth0/auth0-react';
import Sidebar from '../Sidebar/Sidebar';

const ChatPage = () => {
  const { user, getAccessTokenSilently } = useAuth0();
  const location = useLocation();
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');

  const queryParams = new URLSearchParams(location.search);
  const user1 = queryParams.get('user1');
  const user2 = queryParams.get('user2');

  useEffect(() => {
    const fetchChat = async () => {
      try {
        const response = await fetch(`https://cloudcord.com/message/chat?user1=${user1}&user2=${user2}`);
        if (!response.ok) throw new Error('Failed to fetch chat');
        const data = await response.json();
        setMessages(data.messages ?? []);
      } catch (error) {
        console.error('❌ Error fetching chat:', error);
      }
    };

    if (user1 && user2) {
      fetchChat();
    }
  }, [user1, user2]);

  const handleSend = async () => {
    if (!input.trim()) return; // Prevent empty sends

    try {
      const token = await getAccessTokenSilently();

      const response = await fetch('https://cloudcord.com/message/send', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          sender: user1,
          receiver: user2,
          content: input.trim(),
        }),
      });

      if (!response.ok) throw new Error('Failed to send message');

      const data = await response.json();

      setMessages(prev => [...prev, data.message]);
      setInput('');
    } catch (error) {
      console.error('❌ Error sending message:', error);
    }
  };

  return (
    <div className="chatPage">
      <Sidebar />
        <div className="chatMain">
          <div className="chatBody">
        <div className="chatBody">
          {messages.length === 0 ? (
            <p className="noMessagesText">Start the conversation!</p>
          ) : (
            messages.map((msg, index) => (
          <div
            key={index}
            className={msg.sent_by_user === user1 ? "myMessage" : "otherMessage"}
          >
            {msg.content}
            <span className="timestamp">{new Date(msg.timestamp).toLocaleTimeString()}</span>
          </div>
          ))
        )}
      </div>
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
