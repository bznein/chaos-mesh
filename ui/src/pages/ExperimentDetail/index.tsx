import { Box, Button, Grid, Grow, Modal, Paper } from '@material-ui/core'
import EventsTable, { EventsTableHandles } from 'components/EventsTable'
import React, { useEffect, useRef, useState } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { setAlert, setAlertOpen } from 'slices/globalStatus'
import { useHistory, useParams } from 'react-router-dom'

import { Ace } from 'ace-builds'
import Alert from '@material-ui/lab/Alert'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import ConfirmDialog from 'components/ConfirmDialog'
import { Event } from 'api/events.type'
import ExperimentConfiguration from 'components/ExperimentConfiguration'
import { ExperimentDetail as ExperimentDetailType } from 'api/experiments.type'
import Loading from 'components/Loading'
import NoteOutlinedIcon from '@material-ui/icons/NoteOutlined'
import PaperTop from 'components/PaperTop'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import T from 'components/T'
import YAMLEditor from 'components/YAMLEditor'
import api from 'api'
import genEventsChart from 'lib/d3/eventsChart'
import { getStateofExperiments } from 'slices/experiments'
import { useIntl } from 'react-intl'
import { usePrevious } from 'lib/hooks'
import { useSelector } from 'react-redux'
import yaml from 'js-yaml'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    eventsChart: {
      height: 200,
      margin: theme.spacing(3),
    },
    eventDetailPaper: {
      position: 'absolute',
      top: 0,
      left: 0,
      width: '100%',
      height: '100%',
      overflowY: 'scroll',
    },
    configPaper: {
      position: 'absolute',
      top: '50%',
      left: '50%',
      width: '50vw',
      height: '80vh',
      transform: 'translate(-50%, -50%)',
      [theme.breakpoints.down('sm')]: {
        width: '90vw',
      },
    },
  })
)

export default function ExperimentDetail() {
  const classes = useStyles()

  const intl = useIntl()

  const history = useHistory()
  const { uuid } = useParams<{ uuid: string }>()

  const { theme } = useSelector((state: RootState) => state.settings)
  const dispatch = useStoreDispatch()

  const chartRef = useRef<HTMLDivElement>(null)
  const eventsTableRef = useRef<EventsTableHandles>(null)

  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<ExperimentDetailType>()
  const [events, setEvents] = useState<Event[]>()
  const prevEvents = usePrevious(events)
  const [yamlEditor, setYAMLEditor] = useState<Ace.Editor>()
  const [configOpen, setConfigOpen] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogInfo, setDialogInfo] = useState({
    title: '',
    description: '',
    action: 'delete',
  })

  const fetchExperimentDetail = () => {
    api.experiments
      .detail(uuid)
      .then(({ data }) => setDetail(data))
      .catch(console.log)
  }

  const fetchEvents = () =>
    api.events
      .events()
      .then(({ data }) => setEvents(data.filter((d) => d.experiment_id === uuid)))
      .catch(console.log)
      .finally(() => {
        setLoading(false)
      })

  useEffect(() => {
    fetchExperimentDetail()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (detail) {
      fetchEvents()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [detail])

  useEffect(() => {
    if (prevEvents !== events && prevEvents?.length !== events?.length && events) {
      const chart = chartRef.current!

      genEventsChart({
        root: chart,
        events,
        onSelectEvent: eventsTableRef.current!.onSelectEvent,
        intl,
        theme,
      })
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [events])

  const onModalOpen = () => setConfigOpen(true)
  const onModalClose = () => setConfigOpen(false)

  const handleAction = (action: string) => () => {
    switch (action) {
      case 'delete':
        setDialogInfo({
          title: `${intl.formatMessage({ id: 'archives.single' })} ${detail!.name}?`,
          description: intl.formatMessage({ id: 'experiments.deleteDesc' }),
          action: 'delete',
        })

        break
      case 'pause':
        setDialogInfo({
          title: `${intl.formatMessage({ id: 'common.pause' })} ${detail!.name}?`,
          description: intl.formatMessage({ id: 'experiments.pauseDesc' }),
          action: 'pause',
        })

        break
      case 'start':
        setDialogInfo({
          title: `${intl.formatMessage({ id: 'common.start' })} ${detail!.name}?`,
          description: intl.formatMessage({ id: 'experiments.startDesc' }),
          action: 'start',
        })

        break
      default:
        break
    }

    setDialogOpen(true)
  }

  const handleExperiment = (action: string) => () => {
    let actionFunc: any

    switch (action) {
      case 'delete':
        actionFunc = api.experiments.deleteExperiment

        break
      case 'pause':
        actionFunc = api.experiments.pauseExperiment

        break
      case 'start':
        actionFunc = api.experiments.startExperiment

        break
      default:
        actionFunc = null
    }

    if (actionFunc === null) {
      return
    }

    setDialogOpen(false)

    actionFunc(uuid)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: `common.${action}Successfully` }),
          })
        )
        dispatch(setAlertOpen(true))
        dispatch(getStateofExperiments())

        if (action === 'delete') {
          history.push('/experiments')
        }

        if (action === 'pause' || action === 'start') {
          fetchExperimentDetail()
        }
      })
      .catch(console.log)
  }

  const handleUpdateExperiment = () => {
    const data = yaml.safeLoad(yamlEditor!.getValue())

    api.experiments
      .update(data)
      .then(() => {
        setConfigOpen(false)
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'common.updateSuccessfully' }),
          })
        )
        dispatch(setAlertOpen(true))
        fetchExperimentDetail()
      })
      .catch(console.log)
  }

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <Grid container spacing={6}>
          <Grid item xs={12}>
            <Box display="flex">
              <Box mr={3}>
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<ArchiveOutlinedIcon />}
                  onClick={handleAction('delete')}
                >
                  {T('archives.single')}
                </Button>
              </Box>
              <Box>
                {detail?.status === 'Paused' ? (
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<PlayCircleOutlineIcon />}
                    onClick={handleAction('start')}
                  >
                    {T('common.start')}
                  </Button>
                ) : (
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<PauseCircleOutlineIcon />}
                    onClick={handleAction('pause')}
                  >
                    {T('common.pause')}
                  </Button>
                )}
              </Box>
            </Box>
          </Grid>

          {detail?.failed_message && (
            <Grid item xs={12}>
              <Alert severity="error">
                An error occurred: <b>{detail.failed_message}</b>
              </Alert>
            </Grid>
          )}

          <Grid item xs={12}>
            <Paper variant="outlined">
              <PaperTop title={T('common.configuration')}>
                <Button
                  variant="outlined"
                  size="small"
                  color="primary"
                  startIcon={<NoteOutlinedIcon />}
                  onClick={onModalOpen}
                >
                  {T('common.update')}
                </Button>
              </PaperTop>
              <Box p={3}>{detail && <ExperimentConfiguration experimentDetail={detail} />}</Box>
            </Paper>
          </Grid>

          <Grid item xs={12}>
            <Paper variant="outlined">
              <PaperTop title={T('common.timeline')} />
              <div ref={chartRef} className={classes.eventsChart} />
            </Paper>
          </Grid>

          <Grid item xs={12}>
            {events && <EventsTable ref={eventsTableRef} events={events} detailed />}
          </Grid>
        </Grid>
      </Grow>

      <Modal open={configOpen} onClose={onModalClose}>
        <Paper className={classes.configPaper}>
          {detail && (
            <>
              <PaperTop title={detail.name}>
                <Button variant="outlined" color="primary" size="small" onClick={handleUpdateExperiment}>
                  {T('common.confirm')}
                </Button>
              </PaperTop>
              <YAMLEditor theme={theme} data={yaml.safeDump(detail.yaml)} mountEditor={setYAMLEditor} />
            </>
          )}
        </Paper>
      </Modal>

      <ConfirmDialog
        open={dialogOpen}
        setOpen={setDialogOpen}
        title={dialogInfo.title}
        description={dialogInfo.description}
        handleConfirm={handleExperiment(dialogInfo.action)}
      />

      {loading && <Loading />}
    </>
  )
}
