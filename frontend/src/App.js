import {useState, useEffect} from 'react';
import './App.css';
import './styles/login.css';
import './styles/chat.css';

function App() {
  const [ws, setWs] = useState(null);
  const [message, setMessage] = useState('');
  const [chat, setChat] = useState([]);
  const [isConnected, setIsConnected] = useState(false);
  const [typingStatus, setTypingStatus] = useState('');
  const [inputName, setInputName] = useState('');
  const [userName, setUserName] = useState('');
  const [recipient, setRecipient] = useState('');
  const [room, setRoom] = useState('General'); // Default room
  const [hasSelectedRoom, setHasSelectedRoom] = useState(false);

  useEffect(() => {
    if (!userName) return;

    const BACKEND_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080';
    const HISTORY_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
    const socket = new WebSocket(`${BACKEND_URL}/ws?username=${userName}&room=${room}`);
    socket.onmessage = (event) => {
      try{
        const data = JSON.parse(event.data);
        console.log('Received WebSocket message:', data);
        if(data.type === 'typing') {
          setTypingStatus(`${data.sender} is typing...`);
          setTimeout(() => setTypingStatus(''), 2000); // Clear typing status after 2 seconds
        }
        else if(data.type === 'message' || data.type === 'private') {
          setChat((prevChat) => [...prevChat, data.content]);
        }
      }
      catch(err){
        console.error('Error parsing WebSocket message:', err);
        setChat((prevChat) => [...prevChat, event.data]);
      }
    };

    socket.onopen = () => {
      setIsConnected(true);
      console.log('WebSocket connected');
    }

    socket.onclose = () => {
      setIsConnected(false);
      console.log('WebSocket disconnected');
    }

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    }

    setWs(socket); // Save WebSocket instance to state

    if(userName) {
      fetch(`${HISTORY_URL}/history?username=${userName}&room=${room}`)
        .then(response => response.json())
        .then(data => {
          const formattedData = data?.map(msg => `${msg.sender}: ${msg.content}`) || [];
          setChat(formattedData.reverse()); // Show oldest messages first
        })
        .catch(error => {
          console.error('Error fetching message history:', error);
        });
    }
    
    return () => {
      if (socket.readyState === WebSocket.OPEN) {
            socket.close(); // Clean up on unmount
        }
    };
  }, [userName, room])

  const sendMessage  = () => {
    if (!ws || !message.trim()) return;

    const msgObj = {
      type: recipient ? 'private' : 'message',
      content: message,
      sender: userName,
      recipient: recipient,
      room: room,
      timestamp: new Date().toISOString()
    };

    ws.send(JSON.stringify(msgObj));
    setMessage(''); // Clear input field after sending
  }

  const submitUserName = (e) => {
    e.preventDefault();
    if(inputName.trim()) {
      setUserName(inputName.trim());
    }
  }

  if (!userName) {
  return (
    <div className="login-container">
      <div className="login-card">
        <h1 className="login-title">Chit-Chat 💬</h1>
        <h2 className="login-subtitle">Please enter your name to join the chat</h2>
        <form onSubmit={submitUserName} className="login-form">
          <input
            className="login-input"
            value={inputName}
            onChange={e => setInputName(e.target.value)}
            placeholder="Enter your name"
          />
          <button 
            type="submit"
            className="login-button"
            disabled={!inputName.trim()}
          >
            Join Chat
          </button>
        </form>
      </div>
    </div>
  );
}
if (!hasSelectedRoom) {
return (
    <div className="login-container">
        <div className="login-card">
          <h1 className="login-title">Welcome, {userName}! 👋</h1>
          <h2 className="login-subtitle">Select a room to start chatting</h2>
          <div className="room-selection">
            <select 
              value={room} 
              onChange={e => setRoom(e.target.value)}
              className="login-input"
            >
              <option value="General">General</option>
              <option value="Sports">Sports</option>
              <option value="Technology">Technology</option>
              <option value="Music">Music</option>
              <option value="Movies">Movies</option>
              <option value="Travel">Travel</option>
              <option value="Food">Food</option>
              <option value="Art">Art</option>
              <option value="Science">Science</option>
              <option value="History">History</option>
              <option value="Literature">Literature</option>
              <option value="Gaming">Gaming</option>
              <option value="Health">Health</option>
              <option value="Fitness">Fitness</option>
              <option value="Education">Education</option>
              <option value="Business">Business</option>
              <option value="Finance">Finance</option>
              <option value="Politics">Politics</option>
              <option value="Environment">Environment</option>
              <option value="Fashion">Fashion</option>
              <option value="Photography">Photography</option>
              <option value="DIY">DIY</option>
              <option value="Parenting">Parenting</option>
              <option value="Relationships">Relationships</option>
              <option value="Pets">Pets</option>
              <option value="Spirituality">Spirituality</option>
              <option value="Comedy">Comedy</option>
              <option value="Memes">Memes</option>
              <option value="Random">Random</option>
            </select>
            <button 
              className="login-button"
              onClick={() => setHasSelectedRoom(true)}
            >
              Join Room
            </button>
          </div>
        </div>
      </div>
  );
}

return (
  <div className="chat-container">
    <div className="chat-header">
      <h1 className="chat-title">Chit-Chat 💬</h1>
      <p className="chat-status">
        {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
      </p>
    </div>

    <div className="chat-messages">
      {chat.map((msg, i) => (
        <div 
          key={`msg-${i}`}
          className={`message ${msg.user === userName ? 'message-self' : 'message-other'}`}
        >
          {msg}
        </div>
      ))}
      <div className="typing-indicator">
        {typingStatus}
      </div>
    </div>

    <div className="chat-input-container">
      <input
      className="chat-input recipient-input"
      value={recipient}
      onChange={e => setRecipient(e.target.value)}
      placeholder="Send to (leave empty for all)"
      disabled={!isConnected}
      />
      <input
        className="chat-input"
        value={message}
        onChange={e => {
          setMessage(e.target.value);
          if(ws && ws.readyState === WebSocket.OPEN) {
            const typingObj = {
              type: 'typing',
              content: '',
              sender: userName,
              recipient: recipient,
              timestamp: new Date().toISOString()
            };
            ws.send(JSON.stringify(typingObj));
          }
        }}
        placeholder="Type a message..."
        disabled={!isConnected}
      />
      <button 
        className="chat-send-button"
        onClick={sendMessage} 
        disabled={!isConnected}
      >
        Send
      </button>
    </div>
  </div>
);
}

export default App;
