import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Login from "./pages/Login.tsx";
import Register from "./pages/Register.tsx";
import Homepage from "./pages/Homepage.tsx";
import DeleteAccount from "./pages/DeleteAccount.tsx";
import ManagePasskeys from "./pages/ManagePasskeys.tsx";

function App(): React.ReactElement {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/sign-up" element={<Register />} />
        <Route path="/home" element={<Homepage />} />
        <Route path="/passkeys" element={<ManagePasskeys />} />
        <Route path="/delete_account" element={<DeleteAccount />} />
        <Route path="*" element={<div>404 Not Found</div>} />
      </Routes>
    </Router>
  );
}

export default App;
