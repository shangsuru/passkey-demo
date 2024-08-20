import React, { useEffect } from "react";
import { Divider } from "../components/layout/Divider";
import { Link } from "../components/input/Link";
import { Layout } from "../components/layout/Layout";
import { PasswordLogin } from "../components/form/PasswordLogin";
import { PasskeyLogin } from "../components/form/PasskeyLogin";
import { isAuthenticated } from "../utils/shared.ts";
import { useNavigate } from "react-router-dom";

export default function Login(): React.ReactElement {
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated()) {
      navigate("/home");
    }
  });

  return (
    <Layout>
      <PasswordLogin />
      <Divider />
      <PasskeyLogin />

      <div className="mt-10">
        <Link href="#" linkText="Trouble logging in?" />
        <Link href="/sign-up" linkText="Sign up for a new account" />
      </div>
    </Layout>
  );
}
