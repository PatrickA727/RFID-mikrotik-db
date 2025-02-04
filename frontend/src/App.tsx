import { Outlet } from 'react-router-dom';
import { ToastContainer } from 'react-toastify'

function App() {

  return (
    <>
      <ToastContainer></ToastContainer>
      <Outlet></Outlet>
    </>
  )
}

export default App
