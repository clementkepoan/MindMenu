'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  Branch,
  Chatbot,
  Restaurant,
  createBranch,
  createChatbot,
  createRestaurant,
  ensureSessionId,
  queryWithHistory,
  getHealth,
  API_BASE_URL,
  updateChatbot,
} from '@/lib/api';

type Msg = { role: 'user' | 'assistant'; text: string };

export default function DemoPage() {
  // BE Health
  const [health, setHealth] = useState<'unknown' | 'ok' | 'error'>('unknown');
  const [healthMsg, setHealthMsg] = useState('Checking backend connectivity...');
  const supabaseLike = /supabase\.co/i.test(API_BASE_URL);
  async function checkHealth() {
    try {
      const h = await getHealth();
      const ok = h?.status === 'ok';
      setHealth(ok ? 'ok' : 'error');
      setHealthMsg(
        ok
          ? `Backend reachable at ${API_BASE_URL}`
          : `Unexpected health response from ${API_BASE_URL}`
      );
    } catch (e: any) {
      setHealth('error');
      setHealthMsg(
        supabaseLike
          ? `NEXT_PUBLIC_API_BASE_URL is pointing to Supabase (${API_BASE_URL}). Set it to your Go backend, e.g., http://localhost:8080. (${e.message})`
          : `Backend not reachable at ${API_BASE_URL} (${e.message})`
      );
    }
  }
  useEffect(() => {
    checkHealth();
  }, []);

  // Step 1: Restaurant
  const [ownerId, setOwnerId] = useState('');
  const [rName, setRName] = useState('');
  const [rDesc, setRDesc] = useState('');
  const [restaurant, setRestaurant] = useState<Restaurant | null>(null);
  const [creatingRestaurant, setCreatingRestaurant] = useState(false);

  // Step 2: Branch
  const [bName, setBName] = useState('');
  const [bAddress, setBAddress] = useState('');
  const [branch, setBranch] = useState<Branch | null>(null);
  const [creatingBranch, setCreatingBranch] = useState(false);

  // Step 3: Chatbot
  const [chatbot, setChatbot] = useState<Chatbot | null>(null);
  const [creatingChatbot, setCreatingChatbot] = useState(false);
  // Chatbot content fields
  const [hours, setHours] = useState('');
  const [apps, setApps] = useState('');
  const [mains, setMains] = useState('');
  const [desserts, setDesserts] = useState('');
  const [nonAlcoholic, setNonAlcoholic] = useState('');
  const [alcoholic, setAlcoholic] = useState('');
  const [updatingVectors, setUpdatingVectors] = useState(false);
  const [updateMsg, setUpdateMsg] = useState<string>('');

  // Step 4: Chat with history
  const sessionKey = useMemo(
    () => (branch ? `${branch.id}` : 'no-branch'),
    [branch]
  );
  const [sessionId, setSessionId] = useState('');
  const [language, setLanguage] = useState<'en' | 'zh' | 'ja' | 'ko'>('en');
  const [question, setQuestion] = useState('');
  const [chatLoading, setChatLoading] = useState(false);
  const [messages, setMessages] = useState<Msg[]>([]);

  useEffect(() => {
    if (!branch) return;
    const sid = ensureSessionId(sessionKey);
    setSessionId(sid);
  }, [branch, sessionKey]);

  function normalizeAnswer(payload: any): string {
    // Try common fields, fallback to JSON
    return (
      payload?.answer ??
      payload?.response ??
      payload?.message ??
      JSON.stringify(payload)
    );
  }

  function parseList(s: string): string[] {
    return s
      .split(/\r?\n|,/)
      .map((t) => t.trim())
      .filter(Boolean);
  }

  async function handleCreateRestaurant() {
    setCreatingRestaurant(true);
    try {
      const r = await createRestaurant({
        name: rName,
        description: rDesc || undefined,
        owner_id: ownerId,
      });
      setRestaurant(r);
    } catch (e: any) {
      alert(`Create restaurant failed:\n${e.message}`);
    } finally {
      setCreatingRestaurant(false);
    }
  }

  async function handleCreateBranch() {
    if (!restaurant) return;
    setCreatingBranch(true);
    try {
      const b = await createBranch({
        restaurant_id: restaurant.id,
        name: bName,
        address: bAddress || undefined,
      });
      setBranch(b);
    } catch (e: any) {
      alert(`Create branch failed:\n${e.message}`);
    } finally {
      setCreatingBranch(false);
    }
  }

  async function handleCreateChatbot() {
    if (!branch) return;
    setCreatingChatbot(true);
    try {
      const cb = await createChatbot({
        branch_id: branch.id,
        content: {
          menu: {
            appetizers: parseList(apps),
            mains: parseList(mains),
            desserts: parseList(desserts),
            non_alcoholic_drinks: parseList(nonAlcoholic),
            alcoholic_drinks: parseList(alcoholic),
          },
          hours: hours || undefined,
        },
      });
      setChatbot(cb);
    } catch (e: any) {
      alert(`Create chatbot failed:\n${e.message}`);
    } finally {
      setCreatingChatbot(false);
    }
  }

  async function handleUpdateChatbot() {
    if (!branch) return;
    setUpdatingVectors(true);
    setUpdateMsg('');
    try {
      const cb = await updateChatbot({
        branch_id: branch.id,
        content: {
          menu: {
            appetizers: parseList(apps),
            mains: parseList(mains),
            desserts: parseList(desserts),
            non_alcoholic_drinks: parseList(nonAlcoholic),
            alcoholic_drinks: parseList(alcoholic),
          },
          hours: hours || undefined,
        },
      });
      setChatbot(cb);
      setUpdateMsg('Update request sent. If content changed, vectors will be rebuilt.');
    } catch (e: any) {
      setUpdateMsg(`Update failed: ${e.message}`);
    } finally {
      setUpdatingVectors(false);
    }
  }

  async function handleAsk() {
    if (!branch || !sessionId || !question.trim()) return;
    setChatLoading(true);
    const q = question.trim();
    setQuestion('');
    setMessages((m) => [...m, { role: 'user', text: q }]);
    try {
      const res = await queryWithHistory({
        branchId: branch.id,
        question: q,
        session_id: sessionId,
        language,
      });
      const ans = normalizeAnswer(res);
      setMessages((m) => [...m, { role: 'assistant', text: ans }]);
    } catch (e: any) {
      setMessages((m) => [
        ...m,
        { role: 'assistant', text: `Error: ${e.message}` },
      ]);
    } finally {
      setChatLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#0E1012] text-white">
      <div className="max-w-3xl mx-auto p-6 space-y-10">
        <h1 className="text-3xl font-bold">MindMenu Demo</h1>
        <div className="flex items-center gap-3 text-sm p-3 rounded border"
             style={{ borderColor: health === 'ok' ? 'rgba(169,251,215,0.4)' : 'rgba(255,99,99,0.4)', background: 'rgba(255,255,255,0.03)' }}>
          <span
            className="inline-block w-2.5 h-2.5 rounded-full"
            style={{ background: health === 'ok' ? '#34D399' : health === 'error' ? '#EF4444' : '#F59E0B' }}
            aria-label={`backend status: ${health}`}
            title={`backend status: ${health}`}
          />
          <span className="text-white/80">{healthMsg}</span>
          <button
            className="ml-auto px-2 py-1 rounded bg-white/5 hover:bg-white/10 border border-white/10"
            onClick={checkHealth}
          >
            Retry
          </button>
        </div>
        {supabaseLike && (
          <div className="text-xs text-white/70 bg-yellow-500/10 border border-yellow-500/30 rounded p-3">
            Tip: Create <code className="text-white">FE/my-app/.env.local</code> with
            <pre className="mt-2 whitespace-pre-wrap">{`NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`}</pre>
            then restart <code>npm run dev</code>.
          </div>
        )}

        {/* Step 1: Create Restaurant */}
        <section className="space-y-3 border border-white/10 rounded-lg p-4">
          <h2 className="text-xl font-semibold">1) Create Restaurant</h2>
          <div className="grid gap-3">
            <label className="grid gap-1">
              <span className="text-sm text-white/70">Owner ID (UUID)</span>
              <input
                className="bg-black/30 border border-white/10 rounded px-3 py-2"
                placeholder="owner UUID (required by BE/RLS)"
                value={ownerId}
                onChange={(e) => setOwnerId(e.target.value)}
              />
            </label>
            <label className="grid gap-1">
              <span className="text-sm text-white/70">Restaurant Name</span>
              <input
                className="bg-black/30 border border-white/10 rounded px-3 py-2"
                placeholder="e.g., Sakura Bistro"
                value={rName}
                onChange={(e) => setRName(e.target.value)}
              />
            </label>
            <label className="grid gap-1">
              <span className="text-sm text-white/70">Description</span>
              <input
                className="bg-black/30 border border-white/10 rounded px-3 py-2"
                placeholder="optional"
                value={rDesc}
                onChange={(e) => setRDesc(e.target.value)}
              />
            </label>
            <button
              className="mt-2 px-4 py-2 rounded bg-[#A9FBD7]/20 hover:bg-[#A9FBD7]/30 border border-[#A9FBD7]/40"
              disabled={!ownerId || !rName || creatingRestaurant}
              onClick={handleCreateRestaurant}
            >
              {creatingRestaurant ? 'Creating...' : 'Create Restaurant'}
            </button>
          </div>
          {restaurant && (
            <p className="text-sm text-[#A9FBD7]">
              Created Restaurant ID: {restaurant.id}
            </p>
          )}
        </section>

        {/* Step 2: Create Branch */}
        <section className="space-y-3 border border-white/10 rounded-lg p-4">
          <h2 className="text-xl font-semibold">2) Create Branch</h2>
          {!restaurant ? (
            <p className="text-white/60">Create a restaurant first.</p>
          ) : (
            <>
              <p className="text-sm text-white/60">
                Using restaurant_id: <span className="text-[#A9FBD7]">{restaurant.id}</span>
              </p>
              <div className="grid gap-3">
                <label className="grid gap-1">
                  <span className="text-sm text-white/70">Branch Name</span>
                  <input
                    className="bg-black/30 border border-white/10 rounded px-3 py-2"
                    placeholder="e.g., Downtown"
                    value={bName}
                    onChange={(e) => setBName(e.target.value)}
                  />
                </label>
                <label className="grid gap-1">
                  <span className="text-sm text-white/70">Address</span>
                  <input
                    className="bg-black/30 border border-white/10 rounded px-3 py-2"
                    placeholder="optional"
                    value={bAddress}
                    onChange={(e) => setBAddress(e.target.value)}
                  />
                </label>
                <button
                  className="mt-2 px-4 py-2 rounded bg-[#A9FBD7]/20 hover:bg-[#A9FBD7]/30 border border-[#A9FBD7]/40"
                  disabled={!bName || creatingBranch}
                  onClick={handleCreateBranch}
                >
                  {creatingBranch ? 'Creating...' : 'Create Branch'}
                </button>
              </div>
              {branch && (
                <p className="text-sm text-[#A9FBD7]">
                  Created Branch ID: {branch.id}
                </p>
              )}
            </>
          )}
        </section>

        {/* Step 3: Create Chatbot */}
        <section className="space-y-3 border border-white/10 rounded-lg p-4">
          <h2 className="text-xl font-semibold">3) Create Chatbot</h2>
          {!branch ? (
            <p className="text-white/60">Create a branch first.</p>
          ) : (
            <>
              <p className="text-sm text-white/60">
                Using branch_id: <span className="text-[#A9FBD7]">{branch.id}</span>
              </p>
              <div className="grid gap-3">
                <label className="grid gap-1">
                  <span className="text-sm text-white/70">Hours</span>
                  <input
                    className="bg-black/30 border border-white/10 rounded px-3 py-2"
                    placeholder="e.g., Monday–Sunday: 9AM–11PM"
                    value={hours}
                    onChange={(e) => setHours(e.target.value)}
                  />
                </label>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  <label className="grid gap-1">
                    <span className="text-sm text-white/70">Appetizers (one per line)</span>
                    <textarea
                      className="bg-black/30 border border-white/10 rounded px-3 py-2 h-24"
                      placeholder={`Spring Rolls - $8\nChicken Wings - $12`}
                      value={apps}
                      onChange={(e) => setApps(e.target.value)}
                    />
                  </label>
                  <label className="grid gap-1">
                    <span className="text-sm text-white/70">Mains (one per line)</span>
                    <textarea
                      className="bg-black/30 border border-white/10 rounded px-3 py-2 h-24"
                      placeholder={`Grilled Salmon - $24\nBeef Burger - $16`}
                      value={mains}
                      onChange={(e) => setMains(e.target.value)}
                    />
                  </label>
                  <label className="grid gap-1">
                    <span className="text-sm text-white/70">Desserts (one per line)</span>
                    <textarea
                      className="bg-black/30 border border-white/10 rounded px-3 py-2 h-24"
                      placeholder={`Cheesecake - $8\nChocolate Brownie - $7`}
                      value={desserts}
                      onChange={(e) => setDesserts(e.target.value)}
                    />
                  </label>
                  <label className="grid gap-1">
                    <span className="text-sm text-white/70">Non‑alcoholic Drinks (one per line)</span>
                    <textarea
                      className="bg-black/30 border border-white/10 rounded px-3 py-2 h-24"
                      placeholder={`Iced Tea - $4\nOrange Juice - $5`}
                      value={nonAlcoholic}
                      onChange={(e) => setNonAlcoholic(e.target.value)}
                    />
                  </label>
                  <label className="grid gap-1">
                    <span className="text-sm text-white/70">Alcoholic Drinks (one per line)</span>
                    <textarea
                      className="bg-black/30 border border-white/10 rounded px-3 py-2 h-24"
                      placeholder={`House Red Wine - $9\nDraft Beer - $7`}
                      value={alcoholic}
                      onChange={(e) => setAlcoholic(e.target.value)}
                    />
                  </label>
                </div>
              </div>
              <button
                className="px-4 py-2 rounded bg-[#A9FBD7]/20 hover:bg-[#A9FBD7]/30 border border-[#A9FBD7]/40"
                disabled={creatingChatbot}
                onClick={handleCreateChatbot}
              >
                {creatingChatbot ? 'Creating...' : 'Create Chatbot'}
              </button>
              <button
                className="ml-2 px-4 py-2 rounded bg-white/10 hover:bg-white/20 border border-white/20"
                disabled={updatingVectors}
                onClick={handleUpdateChatbot}
                title="Send updated content to trigger vector reindex"
              >
                {updatingVectors ? 'Updating...' : 'Update Chatbot Vectors'}
              </button>
              {chatbot && (
                <p className="text-sm text-[#A9FBD7]">
                  Chatbot ID: {chatbot.id} | Status: {chatbot.status}
                </p>
              )}
              {updateMsg && (
                <p className="text-xs text-white/70">{updateMsg}</p>
              )}
            </>
          )}
        </section>

        {/* Step 4: History-sensitive Chat */}
        <section className="space-y-3 border border-white/10 rounded-lg p-4">
          <h2 className="text-xl font-semibold">4) Chat (with history)</h2>
          {!branch ? (
            <p className="text-white/60">Create a branch (and optionally chatbot) first.</p>
          ) : (
            <>
              <div className="flex items-center gap-3 text-sm text-white/70">
                <div>
                  Session ID:{' '}
                  <span className="text-[#A9FBD7]">
                    {sessionId || '(creating...)'}
                  </span>
                </div>
                <label className="ml-auto flex items-center gap-2">
                  <span>Language</span>
                  <select
                    className="bg-black/30 border border-white/10 rounded px-2 py-1"
                    value={language}
                    onChange={(e) => setLanguage(e.target.value as any)}
                  >
                    <option value="en">English</option>
                    <option value="zh">中文</option>
                    <option value="ja">日本語</option>
                    <option value="ko">한국어</option>
                  </select>
                </label>
              </div>

              <div className="h-64 overflow-auto bg-black/20 border border-white/10 rounded p-3 space-y-2">
                {messages.length === 0 ? (
                  <div className="text-white/50 text-sm">
                    Ask something about the restaurant once you’re ready.
                  </div>
                ) : (
                  messages.map((m, idx) => (
                    <div
                      key={idx}
                      className={
                        m.role === 'user'
                          ? 'text-right'
                          : 'text-left text-[#A9FBD7]'
                      }
                    >
                      <span className="text-xs uppercase text-white/40 mr-2">
                        {m.role}
                      </span>
                      <span>{m.text}</span>
                    </div>
                  ))
                )}
              </div>

              <div className="flex gap-2">
                <input
                  className="flex-1 bg-black/30 border border-white/10 rounded px-3 py-2"
                  placeholder="Type your question..."
                  value={question}
                  onChange={(e) => setQuestion(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleAsk();
                  }}
                />
                <button
                  className="px-4 py-2 rounded bg-[#A9FBD7]/20 hover:bg-[#A9FBD7]/30 border border-[#A9FBD7]/40"
                  disabled={!question.trim() || chatLoading}
                  onClick={handleAsk}
                >
                  {chatLoading ? 'Asking...' : 'Ask'}
                </button>
              </div>
            </>
          )}
        </section>
      </div>
    </main>
  );
}
