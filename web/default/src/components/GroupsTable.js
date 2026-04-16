import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Input, Modal, Pagination, Segment, Table } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';

const GroupsTable = () => {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [openModal, setOpenModal] = useState(false);
  const [editGroup, setEditGroup] = useState({ id: 0, name: '', ratio: 1 });
  const [isEdit, setIsEdit] = useState(false);

  const ITEMS_PER_PAGE = 10;

  const loadGroups = async () => {
    setLoading(true);
    try {
      const res = await API.get(`/api/group/`);
      const { success, message, data } = res.data;
      if (success) {
        setGroups(data);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  useEffect(() => {
    loadGroups();
  }, []);

  const handlePageChange = (e, { activePage }) => {
    setActivePage(activePage);
  };

  const handleSearchChange = (e, { value }) => {
    setSearchKeyword(value);
    setActivePage(1);
  };

  const handleOpenModal = (group = null) => {
    if (group) {
      setEditGroup({ id: group.id, name: group.name, ratio: group.ratio });
      setIsEdit(true);
    } else {
      setEditGroup({ id: 0, name: '', ratio: 1 });
      setIsEdit(false);
    }
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
  };

  const handleInputChange = (e, { name, value }) => {
    setEditGroup({ ...editGroup, [name]: value });
  };

  const handleSubmit = async () => {
    if (!editGroup.name) {
      showError('分组名称不能为空');
      return;
    }
    try {
      let res;
      if (isEdit) {
        res = await API.put(`/api/group/${editGroup.id}`, {
          name: editGroup.name,
          ratio: parseFloat(editGroup.ratio) || 1,
        });
      } else {
        res = await API.post(`/api/group/`, {
          name: editGroup.name,
          ratio: parseFloat(editGroup.ratio) || 1,
        });
      }
      const { success, message } = res.data;
      if (success) {
        showSuccess(isEdit ? '分组更新成功' : '分组创建成功');
        handleCloseModal();
        loadGroups();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const handleDelete = async (group) => {
    if (group.name === 'default') {
      showError('默认分组不可删除');
      return;
    }
    if (!window.confirm(`确定要删除分组 "${group.name}" 吗？`)) {
      return;
    }
    try {
      const res = await API.delete(`/api/group/${group.id}`);
      const { success, message } = res.data;
      if (success) {
        showSuccess('分组删除成功');
        loadGroups();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const filteredGroups = groups.filter((g) =>
    g.name && g.name.toLowerCase().includes(searchKeyword.toLowerCase())
  );
  const totalPages = Math.ceil(filteredGroups.length / ITEMS_PER_PAGE);
  const displayedGroups = filteredGroups.slice(
    (activePage - 1) * ITEMS_PER_PAGE,
    activePage * ITEMS_PER_PAGE
  );

  return (
    <>
      <Segment>
        <Header as='h3'>分组管理</Header>
        <div style={{ marginBottom: '1rem', display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <Input
            placeholder='搜索分组名称...'
            value={searchKeyword}
            onChange={handleSearchChange}
            icon='search'
          />
          <Button color='green' onClick={() => handleOpenModal()} content='新建分组' />
          <Button onClick={loadGroups} content='刷新' icon='refresh' />
        </div>
        <Table basic='very' selectable>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell>ID</Table.HeaderCell>
              <Table.HeaderCell>名称</Table.HeaderCell>
              <Table.HeaderCell>倍率</Table.HeaderCell>
              <Table.HeaderCell>创建时间</Table.HeaderCell>
              <Table.HeaderCell>更新时间</Table.HeaderCell>
              <Table.HeaderCell>操作</Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {displayedGroups.map((group) => (
              <Table.Row key={group.id}>
                <Table.Cell>{group.id}</Table.Cell>
                <Table.Cell>{group.name}</Table.Cell>
                <Table.Cell>{group.ratio}</Table.Cell>
                <Table.Cell>{group.created_at ? new Date(group.created_at).toLocaleString() : '-'}</Table.Cell>
                <Table.Cell>{group.updated_at ? new Date(group.updated_at).toLocaleString() : '-'}</Table.Cell>
                <Table.Cell>
                  <Button size='small' onClick={() => handleOpenModal(group)} content='编辑' />
                  <Button
                    size='small'
                    color='red'
                    onClick={() => handleDelete(group)}
                    content='删除'
                    disabled={group.name === 'default'}
                  />
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
        {totalPages > 1 && (
          <div style={{ marginTop: '1rem', textAlign: 'center' }}>
            <Pagination
              activePage={activePage}
              onPageChange={handlePageChange}
              totalPages={totalPages}
            />
          </div>
        )}
      </Segment>

      <Modal open={openModal} onClose={handleCloseModal} size='small'>
        <Modal.Header>{isEdit ? '编辑分组' : '新建分组'}</Modal.Header>
        <Modal.Content>
          <Form>
            <Form.Field>
              <label>分组名称</label>
              <Input
                name='name'
                value={editGroup.name}
                onChange={handleInputChange}
                placeholder='输入分组名称'
                disabled={isEdit && editGroup.name === 'default'}
              />
            </Form.Field>
            <Form.Field>
              <label>费率倍率</label>
              <Input
                name='ratio'
                type='number'
                step='0.1'
                min='0'
                value={editGroup.ratio}
                onChange={handleInputChange}
                placeholder='输入费率倍率'
              />
            </Form.Field>
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={handleCloseModal}>取消</Button>
          <Button color='green' onClick={handleSubmit}>
            {isEdit ? '更新' : '创建'}
          </Button>
        </Modal.Actions>
      </Modal>
    </>
  );
};

export default GroupsTable;
