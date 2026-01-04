import { useState, useRef, useEffect } from 'react';
import { SEInput } from '../SEInput';
import { SEButton } from '../SEButton';

export interface InlineEditableProps {
  value: string;
  onSave: (value: string) => Promise<void> | void;
  placeholder?: string;
  className?: string;
  disabled?: boolean;
}

export function InlineEditable({
  value,
  onSave,
  placeholder = 'Click to edit',
  className = '',
  disabled = false,
}: InlineEditableProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editValue, setEditValue] = useState(value);
  const [isSaving, setIsSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  useEffect(() => {
    setEditValue(value);
  }, [value]);

  const handleStartEdit = () => {
    if (disabled) return;
    setIsEditing(true);
    setEditValue(value);
  };

  const handleCancel = () => {
    setIsEditing(false);
    setEditValue(value);
  };

  const handleSave = async () => {
    if (editValue === value) {
      setIsEditing(false);
      return;
    }

    setIsSaving(true);
    try {
      await onSave(editValue);
      setIsEditing(false);
    } catch (error) {
      console.error('Failed to save:', error);
    } finally {
      setIsSaving(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSave();
    } else if (e.key === 'Escape') {
      handleCancel();
    }
  };

  if (isEditing) {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        <SEInput
          ref={inputRef}
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={isSaving}
          className="flex-1"
        />
        <SEButton size="sm" onPress={handleSave} isLoading={isSaving}>
          Save
        </SEButton>
        <SEButton size="sm" variant="light" onPress={handleCancel} disabled={isSaving}>
          Cancel
        </SEButton>
      </div>
    );
  }

  return (
    <div
      onClick={handleStartEdit}
      className={`cursor-pointer hover:bg-[var(--se-surface-hover)] rounded-[var(--se-radius-sm)] p-1 ${className}`}
      title="Click to edit"
    >
      {value || <span className="text-[var(--se-text-muted)]">{placeholder}</span>}
    </div>
  );
}

