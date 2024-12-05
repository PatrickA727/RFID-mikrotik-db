import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, createRoutesFromElements, Route, RouterProvider } from 'react-router-dom';
import App from './App.tsx'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import './index.css'
import ItemTypesScreen from './screens/ItemTypesScreen.tsx';
import HomeScreen from './screens/HomeScreen.tsx';
import SellPage from './screens/SellPage.tsx';

const queryClient = new QueryClient()

const router = createBrowserRouter(
  createRoutesFromElements(
    <Route path='/' element={<App/>}>
      <Route index={true} path='/' element={<HomeScreen/>}/>
      <Route path='/type' element={<ItemTypesScreen/>}/>
      <Route path='/sell' element={<SellPage/>}/>
    </Route>
  )
)

createRoot(document.getElementById('root')!).render(
  <QueryClientProvider client={queryClient}>
    <StrictMode>
      <RouterProvider router={router}></RouterProvider>
    </StrictMode>
  </QueryClientProvider>
)
