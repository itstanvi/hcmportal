import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { extractError } from '../api/client';

export default function Login() {
  const [email, setEmail] = useState('admin@acme.com');
  const [password, setPassword] = useState('Password@123');
  const [error, setError] = useState('');
  const { login, loading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = location.state?.from?.pathname || '/employees';

  const onSubmit = async (e) => {
    e.preventDefault();
    setError('');
    try {
      await login(email, password);
      navigate(from, { replace: true });
    } catch (err) {
      setError(extractError(err));
    }
  };

  return (
    <div className="login-screen">
      <div className="card login-card">
        <h1>Sign in to HCM</h1>
        <p>Use your tenant credentials to continue.</p>
        {error && <div className="alert error">{error}</div>}
        <form onSubmit={onSubmit}>
          <div className="mb-12">
            <label>Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="mb-12">
            <label>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <button type="submit" className="primary" disabled={loading} style={{ width: '100%' }}>
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
        <p className="muted mt-16" style={{ fontSize: 12 }}>
          Demo accounts (after running seed):<br />
          <code>admin@acme.com</code> / <code>Password@123</code> (ADMIN)<br />
          <code>hr@acme.com</code> / <code>Password@123</code> (HR)
        </p>
      </div>
    </div>
  );
}
