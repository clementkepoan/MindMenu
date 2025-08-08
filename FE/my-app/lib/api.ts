export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

const isSupabaseBase = /supabase\.co/i.test(API_BASE_URL);

export async function getHealth(): Promise<{ status: string }> {
  if (isSupabaseBase) {
    throw new Error(
      'Misconfigured NEXT_PUBLIC_API_BASE_URL: it points to a Supabase domain. Set it to your Go backend base URL (e.g., http://localhost:8080).'
    );
  }
  const res = await fetch(`${API_BASE_URL}/health`, { cache: 'no-store' });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`Health check failed: ${res.status} ${text}`);
  }
  return res.json();
}

async function postJSON<T>(path: string, body: any): Promise<T> {
  try {
    if (isSupabaseBase) {
      throw new Error(
        'Misconfigured NEXT_PUBLIC_API_BASE_URL: it points to a Supabase domain. Set it to your Go backend base URL (e.g., http://localhost:8080).'
      );
    }
    const res = await fetch(`${API_BASE_URL}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Request failed: ${res.status} ${text}`);
    }
    return res.json() as Promise<T>;
  } catch (err: any) {
    const hint =
      err?.message?.includes('Failed to fetch')
        ? ' Network error. Ensure the Go backend is running, URL is correct, and CORS allows http://localhost:3000.'
        : '';
    throw new Error(`${err?.message || err}${hint ? ' —' + hint : ''}`);
  }
}

export type Restaurant = {
  id: string;
  name: string;
  description?: string;
  owner_id: string;
};

export type Branch = {
  id: string;
  restaurant_id: string;
  name: string;
  address?: string;
  has_chatbot?: boolean;
};

export type Chatbot = {
  id: string;
  branch_id: string;
  status: 'active' | 'building' | 'error' | string;
  content_hash?: string;
};

export type ChatbotContent = {
  menu: {
    appetizers?: string[];
    mains?: string[];
    desserts?: string[];
    non_alcoholic_drinks?: string[];
    alcoholic_drinks?: string[];
  };
  hours?: string;
  servicefee?: string;
};

export async function createRestaurant(input: {
  name: string;
  description?: string;
  owner_id: string;
}): Promise<Restaurant> {
  return postJSON<Restaurant>('/restaurants', input);
}

export async function createBranch(input: {
  restaurant_id: string;
  name: string;
  address?: string;
}): Promise<Branch> {
  return postJSON<Branch>('/branches', input);
}

export async function createChatbot(input: {
  branch_id: string;
  content: ChatbotContent;
}): Promise<Chatbot> {
  return postJSON<Chatbot>('/chatbots', input);
}

// Alias for updating chatbot vectors/content. Backend re-builds if content_hash changes.
export async function updateChatbot(input: {
  branch_id: string;
  content: ChatbotContent;
}): Promise<Chatbot> {
  return postJSON<Chatbot>('/chatbots', input);
}

export async function queryWithHistory(params: {
  branchId: string;
  question: string;
  session_id: string;
  language?: 'en' | 'zh' | 'ja' | 'ko';
}): Promise<any> {
  const { branchId, ...body } = params;
  return postJSON<any>(`/branches/${branchId}/query-with-history`, body);
}

export function ensureSessionId(key: string): string {
  if (typeof window === 'undefined') return '';
  const storageKey = `mindmenu_session_${key}`;
  let v = localStorage.getItem(storageKey);
  if (!v) {
    v = crypto.randomUUID();
    localStorage.setItem(storageKey, v);
  }
  return v;
}
