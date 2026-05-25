import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api, { extractError } from '../api/client';

const STATUS_OPTIONS = ['ACTIVE', 'INACTIVE', 'ON_LEAVE', 'TERMINATED'];

export default function EmployeeList() {
  const navigate = useNavigate();
  const [items, setItems] = useState([]);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [search, setSearch] = useState('');
  const [department, setDepartment] = useState('');
  const [status, setStatus] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchData = async () => {
    setLoading(true);
    setError('');
    try {
      const res = await api.get('/employees', {
        params: { page, page_size: pageSize, search, department, status },
      });
      setItems(res.data.items || []);
      setTotal(res.data.total);
      setTotalPages(res.data.total_pages || 1);
    } catch (err) {
      setError(extractError(err));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, [page]); // eslint-disable-line

  const onSearch = (e) => {
    e.preventDefault();
    setPage(1);
    fetchData();
  };

  const statusBadge = (s) => {
    const cls = s === 'ACTIVE' ? 'success' : s === 'TERMINATED' ? 'danger' : s === 'ON_LEAVE' ? 'warning' : 'muted';
    return <span className={`badge ${cls}`}>{s}</span>;
  };

  return (
    <>
      <div className="page-header">
        <h2>Employees</h2>
        <button className="primary" onClick={() => navigate('/employees/new')}>+ Add Employee</button>
      </div>

      {error && <div className="alert error">{error}</div>}

      <form className="toolbar" onSubmit={onSearch}>
        <input
          placeholder="Search by name, email, employee code"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <input
          placeholder="Department"
          value={department}
          onChange={(e) => setDepartment(e.target.value)}
        />
        <select value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">All statuses</option>
          {STATUS_OPTIONS.map((s) => <option key={s} value={s}>{s}</option>)}
        </select>
        <button type="submit" className="primary">Apply</button>
        <button
          type="button"
          onClick={() => { setSearch(''); setDepartment(''); setStatus(''); setPage(1); setTimeout(fetchData, 0); }}
        >
          Reset
        </button>
      </form>

      {loading ? (
        <div className="empty">Loading…</div>
      ) : items.length === 0 ? (
        <div className="empty">No employees found.</div>
      ) : (
        <>
          <table>
            <thead>
              <tr>
                <th>Code</th>
                <th>Name</th>
                <th>Email</th>
                <th>Department</th>
                <th>Designation</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {items.map((e) => (
                <tr key={e.id} onClick={() => navigate(`/employees/${e.id}`)}>
                  <td>{e.employee_code}</td>
                  <td>{e.name}</td>
                  <td>{e.email}</td>
                  <td>{e.department || '—'}</td>
                  <td>{e.designation || '—'}</td>
                  <td>{statusBadge(e.status)}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="pagination">
            <div className="info">
              Showing page {page} of {totalPages} · {total} total
            </div>
            <div className="controls">
              <button disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</button>
              <button disabled={page >= totalPages} onClick={() => setPage(page + 1)}>Next</button>
            </div>
          </div>
        </>
      )}
    </>
  );
}
