import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Login from "./pages/Login.tsx";
import Register from "./pages/Register.tsx";
import Homepage from "./pages/Homepage.tsx";
import ManagePasskeys from "./pages/ManagePasskeys.tsx";

function App(): React.ReactElement {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/sign-up" element={<Register />} />
        <Route path="/home" element={<Homepage />} />
        <Route path="/passkeys" element={<ManagePasskeys />} />
      </Routes>
    </Router>
  );
}

export default App;
