// DynamicField renders a single form input based on backend field configuration.
// All validation rules are derived from the field config — nothing is hardcoded
// to a specific tenant.

const emailRe = /^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$/;
const phoneRe = /^\+?[0-9\s\-]{7,20}$/;

export function buildFieldRules(field) {
  const rules = {};
  if (field.required) {
    rules.required = `${field.field_name} is required`;
  }
  switch (field.field_type) {
    case 'email':
      rules.pattern = { value: emailRe, message: `${field.field_name} must be a valid email` };
      break;
    case 'phone':
      rules.pattern = { value: phoneRe, message: `${field.field_name} must be a valid phone number` };
      break;
    case 'number':
      rules.validate = (v) => {
        if (v === '' || v === null || v === undefined) return true;
        return !isNaN(Number(v)) || `${field.field_name} must be a number`;
      };
      break;
    default:
      break;
  }
  return rules;
}

export default function DynamicField({ field, register, errors }) {
  const name = `custom_fields.${field.field_key}`;
  const error = errors?.custom_fields?.[field.field_key];
  const rules = buildFieldRules(field);

  let input;
  switch (field.field_type) {
    case 'dropdown': {
      let opts = [];
      if (Array.isArray(field.options)) opts = field.options;
      else if (typeof field.options === 'string') {
        try { opts = JSON.parse(field.options); } catch { opts = []; }
      }
      input = (
        <select {...register(name, rules)} defaultValue="">
          <option value="">— Select —</option>
          {opts.map((o) => <option key={o} value={o}>{o}</option>)}
        </select>
      );
      break;
    }
    case 'boolean':
      input = (
        <select {...register(name, rules)} defaultValue="">
          <option value="">—</option>
          <option value="true">Yes</option>
          <option value="false">No</option>
        </select>
      );
      break;
    case 'date':
      input = <input type="date" {...register(name, rules)} />;
      break;
    case 'number':
      input = <input type="number" step="any" {...register(name, rules)} />;
      break;
    case 'email':
      input = <input type="email" {...register(name, rules)} />;
      break;
    case 'phone':
      input = <input type="tel" {...register(name, rules)} />;
      break;
    case 'text':
    default:
      input = <input type="text" {...register(name, rules)} />;
  }

  return (
    <div>
      <label>
        {field.field_name}
        {field.required && <span className="required-mark">*</span>}
      </label>
      {input}
      {error && <div className="field-error">{error.message}</div>}
    </div>
  );
}
