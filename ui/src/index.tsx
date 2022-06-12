import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './index.scss'
import "@fontsource/lato";
import "@fontsource/inter";

ReactDOM.createRoot(
  document.getElementById('root') as Element
).render(
   <React.StrictMode>
     <App/>
   </React.StrictMode>
)
