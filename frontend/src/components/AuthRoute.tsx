import { Navigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import axios from 'axios';

interface AuthContextType {
    isAuthenticated: boolean;
    isLoading: boolean;
  }

const AuthRoute = ({ children }) => {
    const [auth, setAuth] = useState<AuthContextType>({
        isAuthenticated: false,
        isLoading: true
    })

    useEffect(() => { 
        const validateToken = async () => {
          try {
            const response = await axios.get('/api/user/auth-client', {
              withCredentials: true // Important for sending cookies
            });
    
            if (response.status < 300 || response.status > 199) {
                setAuth({
                  isAuthenticated: true,
                  isLoading: false
                });
            }

          } catch { 
            setAuth({
              isAuthenticated: false,
              isLoading: false
            });
          }
        };
    
        validateToken();
      }, []);   // This useEffect is triggered on component mount and only once

      if (auth.isLoading) {
        return <div>Loading...</div>;
      }

      return auth.isAuthenticated ? <>{children}</>  : <Navigate to="/" />;
}

export default AuthRoute