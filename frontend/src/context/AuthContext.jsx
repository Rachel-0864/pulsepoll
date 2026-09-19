import { createContext, useContext, useEffect, useState } from "react";
import { api } from "../api/api";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);

  useEffect(() => {
    try {
      const savedUser = localStorage.getItem("pulsepoll_user");

      if (savedUser) {
        setUser(JSON.parse(savedUser));
      }
    } catch (error) {
      console.error("Failed to load saved user:", error);
      localStorage.removeItem("pulsepoll_user");
    }
  }, []);

  const login = async (email, password) => {
    const response = await api.login({
      email,
      password,
    });

    const data = response.data;

    // Backend returns:
    // { token, name, email }

    const loggedInUser = {
      name: data.name,
      email: data.email,
      token: data.token,
    };

    localStorage.setItem(
      "pulsepoll_user",
      JSON.stringify(loggedInUser)
    );

    setUser(loggedInUser);

    return data;
  };

  const signup = async (name, email, password) => {
    const response = await api.signup({
      name,
      email,
      password,
    });

    const data = response.data;

    // Backend returns:
    // { token, name, email }

    const signedUpUser = {
      name: data.name,
      email: data.email,
      token: data.token,
    };

    localStorage.setItem(
      "pulsepoll_user",
      JSON.stringify(signedUpUser)
    );

    setUser(signedUpUser);

    return data;
  };

  const logout = () => {
    localStorage.removeItem("pulsepoll_user");
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token: user?.token || null,
        isAuthenticated: !!user?.token,
        login,
        signup,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}