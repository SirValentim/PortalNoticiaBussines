export interface Post {
  id: number;
  title: string;
  slug: string;
  excerpt: string;
  content: string;
  coverImageKey?: string;
  categoryId?: number;
  categoryName?: string;
  authorName?: string;
  status: 'draft' | 'scheduled' | 'published' | 'archived';
  isSponsored: boolean;
  editorialNotes?: string;
  editorResponsible?: string;
  publishedAt?: string;
  createdAt: string;
}

export interface Category {
  id: number;
  slug: string;
  name: string;
  description?: string;
  requiresEditorialNotes: boolean;
}

export interface Store {
  id: number;
  slug: string;
  name: string;
  description?: string;
  category?: string;
  address?: string;
  phone?: string;
  whatsapp?: string;
  logoKey?: string;
  coverImageKey?: string;
  isSponsored: boolean;
  isFeatured: boolean;
  active: boolean;
  createdAt: string;
}

export interface Promotion {
  id: number;
  storeId: number;
  storeName?: string;
  storeSlug?: string;
  title: string;
  slug: string;
  description?: string;
  priceDisplay?: string;
  imageKey?: string;
  startDate: string;
  endDate: string;
  status: string;
  isSponsored: boolean;
}

export interface Banner {
  id: number;
  name: string;
  position: 'hero' | 'sidebar_top' | 'sidebar_bottom' | 'in_feed' | 'sticky_footer';
  imageKey: string;
  linkUrl: string;
  startDate: string;
  endDate: string;
  active: boolean;
  priority: number;
}

export interface Neighborhood {
  id: number;
  slug: string;
  name: string;
  description?: string;
  metaTitle?: string;
  metaDescription?: string;
  coverImageKey?: string;
}

export interface User {
  id: number;
  name: string;
  email: string;
  role: 'admin' | 'editor' | 'comercial';
  active: boolean;
}

export interface Metric {
  id: number;
  metricType: string;
  entityType: string;
  entityId: number;
  createdAt: string;
}

export interface SEOData {
  title: string;
  description: string;
  image?: string;
  url?: string;
  type?: string;
  noIndex?: boolean;
  publishedAt?: string;
  modifiedAt?: string;
  author?: string;
}
