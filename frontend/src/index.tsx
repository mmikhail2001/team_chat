import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';

async function login() {
  console.log("async function login")
  try {
    const responseCheck = await fetch('/api/checkLogin', {
      method: 'POST',
    });
    if (responseCheck.status !== 200) {
      const responseLogin = await fetch('/api/login', {
        method: 'POST',
      });
      
      if (responseLogin.status !== 200) {
        console.error('Server responded with error:', responseLogin.status);
      } else {
        const data = await responseLogin.json();
        if (data.redirect) {
          window.location.href = data.redirect;
        }
      }
    }
  } catch (error) {
    console.error('Error during fetch:', error);
  }
}

login().then(() => {
  console.log("login then")
  ReactDOM.render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
    document.getElementById('root')
  );
});
