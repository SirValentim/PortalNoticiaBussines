import { Routes, Route } from 'react-router';
import { AppProvider } from '@/hooks/useAppStore';
import { Toaster } from '@/components/ui/sonner';
import Layout from '@/sections/Layout';
import Home from '@/pages/Home';
import PostDetail from '@/pages/PostDetail';
import CategoryPosts from '@/pages/CategoryPosts';
import StoreList from '@/pages/StoreList';
import StoreDetail from '@/pages/StoreDetail';
import PromoList from '@/pages/PromoList';
import PromoDetail from '@/pages/PromoDetail';
import NeighborhoodDetail from '@/pages/NeighborhoodDetail';
import Search from '@/pages/Search';
import About from '@/pages/About';
import Contact from '@/pages/Contact';
import Login from '@/pages/Login';
import Admin from '@/pages/Admin';

function App() {
  return (
    <AppProvider>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Home />} />
          <Route path="/noticia/:slug" element={<PostDetail />} />
          <Route path="/categoria/:slug" element={<CategoryPosts />} />
          <Route path="/lojas" element={<StoreList />} />
          <Route path="/loja/:slug" element={<StoreDetail />} />
          <Route path="/promocoes" element={<PromoList />} />
          <Route path="/promocao/:slug" element={<PromoDetail />} />
          <Route path="/bairro/:slug" element={<NeighborhoodDetail />} />
          <Route path="/busca" element={<Search />} />
          <Route path="/sobre" element={<About />} />
          <Route path="/contato" element={<Contact />} />
          <Route path="/login" element={<Login />} />
          <Route path="/painel/7x9k2m" element={<Admin />} />
          <Route path="/painel/7x9k2m/:section" element={<Admin />} />
        </Route>
      </Routes>
      <Toaster />
    </AppProvider>
  );
}

export default App;
