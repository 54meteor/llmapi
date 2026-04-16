import React, { useEffect, useState } from 'react';
import { Button, Form, Modal } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';

const ChannelSelector = ({ open, onClose, onSuccess, tokenId }) => {
  const [channels, setChannels] = useState([]);
  const [selectedChannelId, setSelectedChannelId] = useState(null);
  const [priority, setPriority] = useState(1);
  const [quotaLimit, setQuotaLimit] = useState(0);
  const [loading, setLoading] = useState(false);

  const loadAvailableChannels = async () => {
    try {
      let res = await API.get('/api/channel/?p=0');
      if (res.data.success) {
        let boundRes = await API.get(`/api/token-channel/${tokenId}`);
        let boundIds = [];
        if (boundRes.data.success) {
          boundIds = boundRes.data.data.map(tc => tc.channel_id);
        }
        setChannels(res.data.data.filter(c => !boundIds.includes(c.id)));
      }
    } catch (error) {
      showError(error.message);
    }
  };

  useEffect(() => {
    if (open) {
      loadAvailableChannels();
    }
  }, [open, tokenId]);

  const handleSubmit = async () => {
    if (!selectedChannelId) {
      showError('请选择渠道');
      return;
    }
    setLoading(true);
    let res = await API.post('/api/token-channel/', {
      token_id: parseInt(tokenId),
      channel_id: selectedChannelId,
      priority: parseInt(priority),
      quota_limit: parseInt(quotaLimit)
    });
    const { success, message } = res.data;
    setLoading(false);
    if (success) {
      showSuccess('绑定成功');
      onSuccess();
    } else {
      showError(message);
    }
  };

  const channelOptions = channels.map(c => ({
    key: c.id,
    text: `${c.name}`,
    value: c.id
  }));

  return (
    <Modal open={open} onClose={onClose}>
      <Modal.Header>绑定渠道</Modal.Header>
      <Modal.Content>
        <Form>
          <Form.Select
            label='选择渠道'
            placeholder='请选择'
            options={channelOptions}
            onChange={(e, { value }) => setSelectedChannelId(value)}
            value={selectedChannelId}
          />
          <Form.Input
            label='优先级（数字越小优先级越高）'
            type='number'
            value={priority}
            onChange={(e, { value }) => setPriority(value)}
          />
          <Form.Input
            label='额度上限（0表示不限）'
            type='number'
            value={quotaLimit}
            onChange={(e, { value }) => setQuotaLimit(value)}
          />
        </Form>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={onClose}>取消</Button>
        <Button color='green' onClick={handleSubmit} loading={loading}>绑定</Button>
      </Modal.Actions>
    </Modal>
  );
};

export default ChannelSelector;