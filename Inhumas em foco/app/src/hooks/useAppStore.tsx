import React, { createContext, useContext, useState, useCallback } from 'react';
import type { Post, Store, Promotion, Banner, Neighborhood, User, Category } from '@/types';
import { posts as initialPosts, stores as initialStores, promotions as initialPromotions, banners as initialBanners, neighborhoods as initialNeighborhoods, users as initialUsers, categories as initialCategories } from '@/data/mock';

interface AppState {
  posts: Post[];
  stores: Store[];
  promotions: Promotion[];
  banners: Banner[];
  neighborhoods: Neighborhood[];
  users: User[];
  categories: Category[];
  currentUser: User | null;
  addPost: (post: Post) => void;
  updatePost: (post: Post) => void;
  deletePost: (id: number) => void;
  addStore: (store: Store) => void;
  updateStore: (store: Store) => void;
  deleteStore: (id: number) => void;
  addPromotion: (promo: Promotion) => void;
  deletePromotion: (id: number) => void;
  addBanner: (banner: Banner) => void;
  deleteBanner: (id: number) => void;
  addNeighborhood: (n: Neighborhood) => void;
  deleteNeighborhood: (id: number) => void;
  addUser: (user: User) => void;
  login: (email: string, password: string) => User | null;
  logout: () => void;
}

const AppContext = createContext<AppState | null>(null);

export function AppProvider({ children }: { children: React.ReactNode }) {
  const [posts, setPosts] = useState<Post[]>(initialPosts);
  const [stores, setStores] = useState<Store[]>(initialStores);
  const [promotions, setPromotions] = useState<Promotion[]>(initialPromotions);
  const [banners, setBanners] = useState<Banner[]>(initialBanners);
  const [neighborhoods, setNeighborhoods] = useState<Neighborhood[]>(initialNeighborhoods);
  const [users, setUsers] = useState<User[]>(initialUsers);
  const [categories] = useState<Category[]>(initialCategories);
  const [currentUser, setCurrentUser] = useState<User | null>(null);

  const addPost = useCallback((post: Post) => setPosts(prev => [post, ...prev]), []);
  const updatePost = useCallback((post: Post) => setPosts(prev => prev.map(p => p.id === post.id ? post : p)), []);
  const deletePost = useCallback((id: number) => setPosts(prev => prev.filter(p => p.id !== id)), []);

  const addStore = useCallback((store: Store) => setStores(prev => [store, ...prev]), []);
  const updateStore = useCallback((store: Store) => setStores(prev => prev.map(s => s.id === store.id ? store : s)), []);
  const deleteStore = useCallback((id: number) => setStores(prev => prev.filter(s => s.id !== id)), []);

  const addPromotion = useCallback((promo: Promotion) => setPromotions(prev => [promo, ...prev]), []);
  const deletePromotion = useCallback((id: number) => setPromotions(prev => prev.filter(p => p.id !== id)), []);

  const addBanner = useCallback((banner: Banner) => setBanners(prev => [banner, ...prev]), []);
  const deleteBanner = useCallback((id: number) => setBanners(prev => prev.filter(b => b.id !== id)), []);

  const addNeighborhood = useCallback((n: Neighborhood) => setNeighborhoods(prev => [...prev, n]), []);
  const deleteNeighborhood = useCallback((id: number) => setNeighborhoods(prev => prev.filter(n => n.id !== id)), []);

  const addUser = useCallback((user: User) => setUsers(prev => [...prev, user]), []);

  const login = useCallback((email: string, password: string) => {
    const user = users.find(u => u.email === email && u.active);
    if (user && password === 'admin123') { // Simplified for demo
      setCurrentUser(user);
      return user;
    }
    return null;
  }, [users]);

  const logout = useCallback(() => setCurrentUser(null), []);

  return (
    <AppContext.Provider value={{
      posts, stores, promotions, banners, neighborhoods, users, categories, currentUser,
      addPost, updatePost, deletePost,
      addStore, updateStore, deleteStore,
      addPromotion, deletePromotion,
      addBanner, deleteBanner,
      addNeighborhood, deleteNeighborhood,
      addUser, login, logout
    }}>
      {children}
    </AppContext.Provider>
  );
}

export function useAppStore() {
  const ctx = useContext(AppContext);
  if (!ctx) throw new Error('useAppStore must be used within AppProvider');
  return ctx;
}
