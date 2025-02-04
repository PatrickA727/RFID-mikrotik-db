import { useMutation } from '@tanstack/react-query';
import axios from 'axios';
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

interface User {
  email: string
  password: string
}

const LoginScreen = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("")

  const navigate = useNavigate()

  useEffect(() => { 
    const validateToken = async () => {
        try {
            const response = await axios.get('/api/user/auth-client', {
                withCredentials: true // Important for sending cookies
            });
            if (response.status < 300 || response.status > 199) {
                navigate("/home");
            } else {
              console.log(response.status)
            }
        } catch(error) { 
            console.log(error)
        }
    };

    validateToken();
    }, []);

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value)
  }

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value)
  }

  const loginUser = async (user: User) => {
    try {
      // console.log("first")
      const response = await axios.post<User>(
        `api/user/login`,
        user
      );
      console.log("res: ", response)
      navigate("/home") 
    } catch (error) {
      console.log("error logging in: ", error)
    };
  }

  const { mutate: loginMutation } = useMutation({
    mutationFn: loginUser,
    onSuccess: () => {
      
    },
    onError: () => {
      console.log("error logging user")
    }
  })

  const loginUserHandler = async (user_email: string, user_password: string) => {
    
    if (user_email !== "" && user_password !== "") {
      const newUser: User = {
        email: user_email,
        password: user_password
      }
      try {
        await loginMutation(newUser)
        setEmail("")
        setPassword("")
      } catch (error) {
        console.log("error log in: ", error)
      }
    } else {
      console.log("field empty")
      return
    }
  }

  return (
      <div className="min-h-screen flex justify-center items-center relative bg-gray-200 pb-20">

        <div className="relative bg-white rounded-xl shadow-xl p-4 max-w-xs w-full h-full">
          <h2 className="text-center text-2xl font-bold text-black mb-4">Login</h2>
          <form
            onSubmit={(event) => {
              event.preventDefault(); // Prevent refresh
              loginUserHandler(email, password); 
            }}
          >
            <div className="mb-4 relative">
              <input
                className="w-full p-2 rounded-lg bg-gray-400 bg-opacity-20 text-black placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                type="text"
                id="email"
                placeholder="Email" 
                value={email ?? ""}
                onChange={handleEmailChange}/>
              <i className="absolute right-3 top-2 text-black font-normal not-italic">👤</i>
            </div>

            <div className="mb-4 relative">
                <input
                  className="w-full p-2 rounded-lg bg-gray-400 bg-opacity-20 text-black placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  type="password"
                  id="password"
                  placeholder="Password" 
                  value={password ?? ""}
                  onChange={handlePasswordChange}/>
                <i className="absolute right-3 top-2 text-white font-normal not-italic">🔑</i>
            </div>

            <button
                className="w-full py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition duration-300"
                type='submit'
            >
                Login
            </button>
        </form>
      </div>
    </div>
  )
}

export default LoginScreen