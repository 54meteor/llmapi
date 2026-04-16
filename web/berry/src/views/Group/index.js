import { useState, useEffect } from 'react';
import { showError, showSuccess, showConfirm } from 'utils/common';

import { useTheme } from '@mui/material/styles';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableContainer from '@mui/material/TableContainer';
import PerfectScrollbar from 'react-perfect-scrollbar';
import TablePagination from '@mui/material/TablePagination';
import LinearProgress from '@mui/material/LinearProgress';

import { Button, Card, Box, Stack, Container, Typography } from '@mui/material';

import GroupTableRow from './component/TableRow';
import GroupTableHead from './component/TableHead';
import TableToolBar from 'ui-component/TableToolBar';
import { API } from 'utils/api';
import { ITEMS_PER_PAGE } from 'constants';
import { IconRefresh, IconPlus } from '@tabler/icons-react';
import EditModal from './component/EditModal';

export default function GroupPage() {
  const [groups, setGroups] = useState([]);
  const [searching, setSearching] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [order, setOrder] = useState('asc');
  const [orderBy, setOrderBy] = useState('id');
  const [openModal, setOpenModal] = useState(false);
  const [editGroupId, setEditGroupId] = useState(0);

  const loadGroups = async () => {
    setSearching(true);
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
    setSearching(false);
  };

  const handleRequestSort = (event, property) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
  };

  const handleSearchKeyword = (event) => {
    setSearchKeyword(event.target.value);
  };

  const searchGroups = async (event) => {
    event.preventDefault();
    if (searchKeyword === '') {
      await loadGroups();
      return;
    }
    await loadGroups();
  };

  const handleEdit = (row) => {
    setEditGroupId(row.id);
    setOpenModal(true);
  };

  const handleDelete = (row) => {
    if (row.name === 'default') {
      showError('默认分组不可删除');
      return;
    }
    showConfirm(`确定要删除分组 "${row.name}" 吗？`, async () => {
      const res = await API.delete(`/api/group/${row.id}`);
      const { success, message } = res.data;
      if (success) {
        showSuccess('分组删除成功！');
        loadGroups();
      } else {
        showError(message);
      }
    });
  };

  const handleOpenModal = (groupId) => {
    setEditGroupId(groupId || 0);
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
    setEditGroupId(0);
  };

  const handleOkModal = (status) => {
    if (status === true) {
      handleCloseModal();
      loadGroups();
    }
  };

  useEffect(() => {
    loadGroups()
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  const filteredGroups = groups.filter(g => g.name && g.name.toLowerCase().includes(searchKeyword.toLowerCase()));
  const sortedGroups = filteredGroups.sort((a, b) => {
    if (order === 'asc') {
      return a[orderBy] > b[orderBy] ? 1 : -1;
    }
    return a[orderBy] < b[orderBy] ? 1 : -1;
  });

  return (
    <>
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={5}>
        <Typography variant="h4">分组</Typography>

        <Button variant="contained" color="primary" startIcon={<IconPlus />} onClick={() => handleOpenModal(0)}>
          新建分组
        </Button>
      </Stack>
      <Card>
        <Box component="form" onSubmit={searchGroups} noValidate>
          <TableToolBar filterName={searchKeyword} handleFilterName={handleSearchKeyword} placeholder={'搜索分组名称 ...'} />
        </Box>
        <Box sx={{ textAlign: 'right', height: 50, display: 'flex', justifyContent: 'space-between', p: (theme) => theme.spacing(0, 1, 0, 3) }}>
          <Container>
            <Button startIcon={<IconRefresh width={'18px'} />} onClick={loadGroups}>
              刷新
            </Button>
          </Container>
        </Box>
        {searching && <LinearProgress />}
        <PerfectScrollbar component="div">
          <TableContainer sx={{ overflow: 'unset' }}>
            <Table sx={{ minWidth: 800 }}>
              <GroupTableHead order={order} orderBy={orderBy} onRequestSort={handleRequestSort} />
              <TableBody>
                {sortedGroups.slice(0, ITEMS_PER_PAGE).map((row) => (
                  <GroupTableRow
                    row={row}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                    key={row.id}
                  />
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </PerfectScrollbar>
        <TablePagination
          component="div"
          count={sortedGroups.length}
          rowsPerPage={ITEMS_PER_PAGE}
          rowsPerPageOptions={[ITEMS_PER_PAGE]}
          page={0}
          onPageChange={() => {}}
        />
      </Card>
      <EditModal open={openModal} onCancel={handleCloseModal} onOk={handleOkModal} groupId={editGroupId} />
    </>
  );
}